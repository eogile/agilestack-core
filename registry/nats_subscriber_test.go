package registry_test

import (
	"testing"
	"time"

	pb "github.com/eogile/agilestack-core/proto"
	"github.com/eogile/agilestack-core/registry"
)

func TestListAvailablePluginsNats(t *testing.T) {
	setUp()

	/*
	 * Initializing the subscriber
	 */
	subscriber := registry.NewNatsSubscriber(localhostNatsServerURL)
	defer subscriber.Shutdown()

	/*
	 * Establishing a connection to publish messages
	 */
	connection := registry.EstablishConnection(localhostNatsServerURL)

	var plugins = pb.Plugins{}
	err := connection.Request(pb.ListAvailablePluginsTopic,
		&pb.Empty{}, &plugins, 5000*time.Millisecond)
	if err != nil {
		t.Errorf("Error should be nil : %v", err)
	}
	if len(plugins.Plugins) == 0 {
		t.Errorf("There should be at least one available plugin")
	}
}

func TestListInstalledPluginsNats(t *testing.T) {
	setUp()

	/*
	 * Initializing the subscriber
	 */
	subscriber := registry.NewNatsSubscriber(localhostNatsServerURL)
	defer subscriber.Shutdown()

	/*
	 * Establishing a connection to publish messages
	 */
	connection := registry.EstablishConnection(localhostNatsServerURL)

	var plugins = pb.Plugins{}
	err := connection.Request(pb.ListInstalledPluginsTopic,
		&pb.Empty{}, &plugins, 5000*time.Millisecond)
	if err != nil {
		t.Errorf("Error should be nil : %v", err)
	}
	if len(plugins.Plugins) > 0 {
		t.Errorf("There should be no installed plugins. got %+v", plugins.Plugins)
	}
}

func TestInstallPluginNats(t *testing.T) {
	setUp()

	/*
	 * Initializing the subscriber
	 */
	subscriber := registry.NewNatsSubscriber(localhostNatsServerURL)
	defer subscriber.Shutdown()

	/*
	 * Establishing a connection to publish messages
	 */
	connection := registry.EstablishConnection(localhostNatsServerURL)

	plugin := &pb.Plugin{Name: "agilestack-room-booking-api"}
	request := pb.InstallPluginRequest{Plugin: plugin}

	var result = pb.NetResponse{}
	err := connection.Request(pb.InstallPluginTopic,
		&request, &result, 10000*time.Millisecond)
	if err != nil {
		t.Errorf("Error should be nil : %v", err)
	}
	if result.Response != pb.Responses_ACK {
		t.Errorf("Invalid response status : %v", result.Response)
	}

	/*
	 * Sleeping to wait the Docker container to be started so it
	 * can register to the registry.
	 */
	time.Sleep(500 * time.Millisecond)

	/*
	 * Checks that the plugin is installed
	 */
	var plugins = pb.Plugins{}
	connection.Request(pb.ListInstalledPluginsTopic,
		&pb.Empty{}, &plugins, 5000*time.Millisecond)

	if len(plugins.Plugins) != 1 {
		t.Errorf("There should be one installed plugin, got %d", len(plugins.Plugins))
		return
	}

	pluginPresent := false
	for _, installedPlugin := range plugins.Plugins {
		if installedPlugin.Name == plugin.Name {
			pluginPresent = true
		}
	}
	if !pluginPresent {
		t.Errorf("Plugin %s is not installed", plugin.Name)
	}
}

func TestUninstallPluginNats(t *testing.T) {
	setUp()

	/*
	 * Initializing the subscriber
	 */
	subscriber := registry.NewNatsSubscriber(localhostNatsServerURL)
	defer subscriber.Shutdown()

	/*
	 * Establishing a connection to publish messages
	 */
	connection := registry.EstablishConnection(localhostNatsServerURL)

	plugin := &pb.Plugin{Name: "agilestack-room-booking-api"}
	request := pb.InstallPluginRequest{Plugin: plugin}

	/*
	 * Installing a plugin
	 */
	var result = pb.NetResponse{}
	connection.Request(pb.InstallPluginTopic,
		&request, &result, 10000*time.Millisecond)

	/*
	 * Sleeping to wait the Docker container to be started so it
	 * can receive SIGTERM signals.
	 */
	time.Sleep(500 * time.Millisecond)

	/*
	 * Uninstalling the plugin
	 */
	err := connection.Request(pb.UninstallPluginTopic,
		plugin, &result, 10000*time.Millisecond)
	if err != nil {
		t.Errorf("Error should be nil : %v", err)
	}
	if result.Response != pb.Responses_ACK {
		t.Errorf("Invalid response status : %v", result.Response)
	}

	/*
	 * Checks that the plugin is not installed
	 */
	var plugins = pb.Plugins{}
	connection.Request(pb.ListInstalledPluginsTopic,
		&pb.Empty{}, &plugins, 10000*time.Millisecond)

	if len(plugins.Plugins) != 0 {
		t.Errorf("There should be no installed plugins")
	}
}
