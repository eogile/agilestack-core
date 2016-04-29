package registry

import (
	"log"

	pb "github.com/eogile/agilestack-core/proto"
)

type Registry interface {

	/*
	 * Returns the list of downloaded plugins that are not registered.
	 */
	ListAvailablePlugins() (*pb.Plugins, error)

	/*
	 * Returns the list of the currently running plugins
	 */
	ListInstalledPlugins() (*pb.Plugins, error)

	/*
	 * Installs the given plugin.
	 *
	 * Please notice that installing a plugin does not mean the plugin is registered.
	 *
	 * If the plugin is already installed, then it's re-installed.
	 */
	InstallPlugin(installRequest pb.InstallPluginRequest) (*pb.NetResponse, error)

	/*
	 * Uninstalls the given plugin.
	 *
	 * Please notice that installing a plugin causes the removal
	 * of the plugin from the list of registered plugins.
	 */
	UninstallPlugin(plugin pb.Plugin) (*pb.NetResponse, error)
}

/*
 * Implementation of "Registry" where all the data is stored
 * in memory.
 */
type InMemoryRegistry struct {
	/*
	 * The client to access the location where the plugins are installed.
	 */
	pluginStorageClient PluginStorageClient
}

func NewInMemoryRegistry(pluginStorageClient PluginStorageClient) *InMemoryRegistry {
	return &InMemoryRegistry{
		pluginStorageClient: pluginStorageClient,
	}
}

func (registry *InMemoryRegistry) ListAvailablePlugins() (*pb.Plugins, error) {
	log.Println("Listing available plugins")
	return registry.pluginStorageClient.ListInstallablePlugins()
}

func (registry *InMemoryRegistry) ListInstalledPlugins() (*pb.Plugins, error) {
	log.Println("Listing installed plugins")
	return registry.pluginStorageClient.ListInstalledPlugins()
}

func (registry *InMemoryRegistry) InstallPlugin(installRequest pb.InstallPluginRequest) (*pb.NetResponse, error) {
	name := installRequest.Plugin.Name
	log.Printf("Installing plugin \"%s\"\n", name)

	/*
	 * First, uninstalling the plugin if required.
	 */
	if isInstalled, _ := registry.pluginStorageClient.IsPluginInstalled(name); isInstalled {
		log.Printf("Plugin \"%s\" is already installed. ", name)
		log.Println("It will be unistalled before installation")
		_, err := registry.UninstallPlugin(pb.Plugin{Name: name})

		if err != nil {
			return nil, err
		}
	}

	err := registry.pluginStorageClient.InstallPlugin(
		name, installRequest.Cmd)
	if err != nil {
		log.Printf("Error while installing the plugin : %v", err)
		return nil, err
	}
	return &pb.NetResponse{Response: pb.Responses_ACK}, nil
}

func (registry *InMemoryRegistry) UninstallPlugin(plugin pb.Plugin) (*pb.NetResponse, error) {
	log.Printf("Uninstalling plugin \"%s\"\n", plugin.Name)

	/*
	 * Uninstalling the plugin.
	 */
	err := registry.pluginStorageClient.UninstallPlugin(plugin.Name)
	if err != nil {
		log.Printf("Error while uninstalling the plugin : %v", err)
		return nil, err
	}
	return &pb.NetResponse{Response: pb.Responses_ACK}, nil
}
