package registry

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/eogile/agilestack-core/proto"
	"github.com/nats-io/nats"
	"github.com/nats-io/nats/encoders/protobuf"
)

type natsSubscriber struct {
	registry      Registry
	connection    *nats.EncodedConn
	pluginFactory pluginFactory

	natsServerURL string
}

func NewNatsSubscriber(natsServerURL string) *natsSubscriber {
	subscriber := &natsSubscriber{}
	subscriber.natsServerURL = natsServerURL

	/*
	 * Initializing the registry
	 */
	dockerWrapper := NewDockerStorageClient()
	subscriber.registry = NewInMemoryRegistry(dockerWrapper)

	connection := EstablishConnection(natsServerURL)
	subscriber.connection = connection

	/*
	 * Initializing the plugins factory.
	 */
	subscriber.pluginFactory = NewPluginFactory()

	/*
	 * Initializing NATS subscriptions
	 */
	subscriber.subscribeToListAvailablePlugins()
	subscriber.subscribeToListInstalledPlugins()
	subscriber.subscribeToInstallPlugin()
	subscriber.subscribeToUninstallPlugin()
	subscriber.subscribeToCreatePlugin()

	return subscriber
}

/*
 * Establishes a connection to the NATS server.
 */
func EstablishConnection(natsServerURL string) *nats.EncodedConn {
	connection, err := nats.Connect(natsServerURL)
	if err != nil {
		log.Fatalf("Error while connecting to the Nats server : %v", err)
	}

	protobufConnection, protobufErr := nats.NewEncodedConn(connection, protobuf.PROTOBUF_ENCODER)
	if protobufErr != nil {
		log.Fatalf("Error while establishing a protobuf connection to the Nats server : %v",
			protobufErr)
	}
	return protobufConnection
}

/*
 * Subscribes to the "listAvailablePlugins" topic.
 */
func (subscriber natsSubscriber) subscribeToListAvailablePlugins() {
	subscriber.connection.Subscribe(pb.ListAvailablePluginsTopic, func(m *nats.Msg) {

		response, _ := subscriber.registry.ListAvailablePlugins()
		subscriber.connection.Publish(m.Reply, response)
	})
}

/*
 * Subscribes to the "listInstalledPlugins" topic.
 */
func (subscriber natsSubscriber) subscribeToListInstalledPlugins() {
	subscriber.connection.Subscribe(pb.ListInstalledPluginsTopic, func(m *nats.Msg) {

		response, _ := subscriber.registry.ListInstalledPlugins()
		subscriber.connection.Publish(m.Reply, response)
	})
}

/*
 * Subscribes to the "installPlugin" topic.
 */
func (subscriber natsSubscriber) subscribeToInstallPlugin() {
	subscriber.connection.Subscribe(pb.InstallPluginTopic, func(_ string, reply string, installRequest *pb.InstallPluginRequest) {

		response, err := subscriber.registry.InstallPlugin(*installRequest)
		if err != nil {
			log.Println("Error while installing the plugin", err)
			subscriber.connection.Publish(reply, &pb.NetResponse{Response: pb.Responses_ERROR, Details: err.Error()})
		} else {
			subscriber.connection.Publish(reply, response)
		}

	})
}

/*
 * Subscribes to the "uninstallPlugin" topic.
 */
func (subscriber natsSubscriber) subscribeToUninstallPlugin() {
	subscriber.connection.Subscribe(pb.UninstallPluginTopic, func(_ string, reply string, plugin *pb.Plugin) {

		response, err := subscriber.registry.UninstallPlugin(*plugin)
		if err != nil {
			log.Println("Error while uninstalling the plugin", err)
			subscriber.connection.Publish(reply, &pb.NetResponse{Response: pb.Responses_ERROR, Details: err.Error()})
		} else {
			log.Printf("Plugin %s was uninstalled.", plugin.Name)
			subscriber.connection.Publish(reply, response)
		}

	})
}

/*
 * Subscribes to the "core.plugin.create" topic.
 *
 * When a message is received on this topic, then a new plugin should be created.
 */
func (subscriber natsSubscriber) subscribeToCreatePlugin() {
	subscriber.connection.Subscribe(pb.CreatePlugin, func(_ string, reply string, request *pb.NewPluginRequest) {
		log.Println("Creating the plugin", request.Name)

		err := subscriber.pluginFactory.CreatePlugin(request)
		status := err == nil
		if err != nil {
			log.Println("Error while creating the plugin.", err)
		}
		log.Printf("Image %s created.\n", request.Name)

		response := &pb.NewPluginResponse{Status: status}
		subscriber.connection.Publish(reply, response)
	})
}

func (subscriber natsSubscriber) Shutdown() {
	subscriber.connection.Close()
}

/*
 * Registering a shutdown hook to stop and remove all plugins
 * containers when the current container stops.
 *
 * This hook is only for dev comfort.
 *
 * FIXME Removes the hook when plugins lifecycle will be properly managed
 */
func (subscriber natsSubscriber) InitShutdownHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	log.Println("Initializing shutdown hook")
	go func() {
		for _ = range c {
			log.Println("Intercepting \"Interrupt\" signal")
			subscriber.stopPlugins()

			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				log.Printf("Error while finding process : %v", err)
				os.Exit(1)
			} else {
				/*
				 * Sending a SIGINT signal.
				 *
				 * Sending a SIGTERM signal does not work.
				 */
				close(c)
				signal.Stop(c)
				process.Signal(os.Interrupt)
			}
		}
	}()
}

/*
 * Stops all the manually started plugins.
 */
func (subscriber natsSubscriber) stopPlugins() {
	plugins, _ := subscriber.registry.ListInstalledPlugins()

	for _, plugin := range plugins.Plugins {
		log.Print("[Shutdown] stopping plugin ", plugin.Name)
		if plugin.Name == "backoffice" {
			continue
		}
		subscriber.registry.UninstallPlugin(*plugin)
	}
}
