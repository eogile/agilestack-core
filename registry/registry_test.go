package registry_test

import (
	"testing"
	"time"

	pb "github.com/eogile/agilestack-core/proto"
	"github.com/eogile/agilestack-core/registry"
	"github.com/fsouza/go-dockerclient"
)

const (
	testPluginName = "agilestack-root-app"
)

var testRegistry *registry.InMemoryRegistry
var dockerWrapper *registry.DockerStorageClient

/*
 * Tests that the list of available plugins contains at least
 * one item.
 */
func TestListAvailablePlugins(t *testing.T) {
	setUp()

	plugins, err := testRegistry.ListAvailablePlugins()
	if err != nil {
		t.Errorf("Error should not be nil : %v", err)
	}

	countAvailablePlugins := len(plugins.Plugins)
	if countAvailablePlugins == 0 {
		t.Fatalf("Expected at least one plugin installable, got zero.")
	}
}

/*
 * Tests that installing a plugin removes it from the list of
 * of available plugin.
 */
func TestListAvailablePluginsWithInstalledPlugins(t *testing.T) {
	setUp()

	/*
	 * Installing a plugin
	 */
	request := pb.InstallPluginRequest{
		Plugin: &pb.Plugin{
			Name: testPluginName,
		},
	}
	_, installErr := testRegistry.InstallPlugin(request)
	if installErr != nil {
		t.Errorf("Error during plugin installation : %v", installErr)
		return
	}

	plugins, err := testRegistry.ListAvailablePlugins()
	if err != nil {
		t.Errorf("Error should not be nil : %v", err)
		return
	}

	for _, plugin := range plugins.Plugins {
		if plugin.Name == testPluginName {
			t.Errorf("The plugin has been installed. It should not be listed as available")
		}
	}
}

/*
 * Tests that, after installing any plugins, the list of
 * installed plugins is empty.
 */
func TestListInstalledPluginsEmptyList(t *testing.T) {
	setUp()

	plugins, err := testRegistry.ListInstalledPlugins()
	if err != nil {
		t.Errorf("Error should not be nil : %v", err)
		return
	}
	if len(plugins.Plugins) != 0 {
		t.Errorf("Invalid number of plugins: should be zero, got %d : %v", len(plugins.Plugins), plugins.Plugins)
	}
}

/*
 * Tests that adding a plugin adds it the list of installed
 * plugins.
 *
 * Tests that un-installing a plugin removes it from the list of
 * installed plugins.
 */
func TestListInstalledPlugins(t *testing.T) {
	setUp()

	request := pb.InstallPluginRequest{
		Plugin: &pb.Plugin{
			Name: testPluginName,
		},
	}
	_, installErr := testRegistry.InstallPlugin(request)
	if installErr != nil {
		t.Errorf("Error during plugin installation : %v", installErr)
		return
	}

	plugins, err := testRegistry.ListInstalledPlugins()
	if err != nil {
		t.Errorf("Error should not be nil : %v", err)
		return
	}

	if !pluginsArrayContains(plugins.Plugins, testPluginName) {
		t.Errorf("The plugin is not present in the list. Got %v", plugins.Plugins)
		return
	}

	/*
	 * Uninstalling the plugin
	 */
	testRegistry.UninstallPlugin(pb.Plugin{Name: testPluginName})

	finalPlugins, finalErr := testRegistry.ListInstalledPlugins()
	if finalErr != nil {
		t.Errorf("Error should not be nil : %v", finalErr)
		return
	}

	if pluginsArrayContains(finalPlugins.Plugins, testPluginName) {
		t.Error("The plugin is present in the list. Got %v", finalPlugins.Plugins)
		return
	}
}

/*
 * Tests the installation of a plugin.
 */
func TestInstallPlugin(t *testing.T) {
	setUp()

	plugin := &pb.Plugin{Name: testPluginName}
	request := pb.InstallPluginRequest{Plugin: plugin}

	response, err := testRegistry.InstallPlugin(request)
	if err != nil {
		t.Errorf("Error during plugin installation : %v", err)
		return
	}
	if response.Response != pb.Responses_ACK {
		t.Errorf("Response status is not ACK : %v", response.Response)
	}

	/*
	 * Checks that the plugin is in the list of installed
	 * plugins
	 */
	assertThatPluginsIsInstalled(t, testPluginName, true)

	/*
	 * Checks that the plugin is not in the list of available
	 * plugins.
	 */
	assertThatPluginsIsAvailable(t, testPluginName, false)
}

/*
 * Test that checks that a plugin installation succeeds even if there is a
 * stopped container that represents this plugin.
 *
 */
func TestInstallPluginWithStoppedContainer(t *testing.T) {
	setUp()

	/*
	 * Container configuration
	 */
	containerConfig := docker.Config{
		Image: testPluginName,
	}

	/*
	 * Host configuration
	 */
	hostConfig := docker.HostConfig{
		PublishAllPorts: true,
		NetworkMode:     "agilestacknet",
	}

	containerOptions := docker.CreateContainerOptions{
		Name:       testPluginName,
		Config:     &containerConfig,
		HostConfig: &hostConfig,
	}

	container, err := dockerClient.CreateContainer(containerOptions)

	if err != nil {
		t.Fatal("Error while creating container", err)
	}
	if container.State.Running {
		t.Fatal("Container should be stopped")
	}

	TestInstallPlugin(t)
}

/*
 * Tests the un-installation of a plugin.
 */
func TestUninstallPlugin(t *testing.T) {
	setUp()
	/*
	 * Installing a plugin
	 */
	plugin := &pb.Plugin{Name: testPluginName}
	request := pb.InstallPluginRequest{Plugin: plugin}
	testRegistry.InstallPlugin(request)

	/*
	 * Sleeping to wait the Docker container to be started so it
	 * can receive SIGTERM signals.
	 */
	time.Sleep(500 * time.Millisecond)

	response, err := testRegistry.UninstallPlugin(*plugin)
	if err != nil {
		t.Errorf("Error during plugin un-installation : %v", err)
		return
	}
	if response.Response != pb.Responses_ACK {
		t.Errorf("Response status is not ACK : %v", response.Response)
	}

	/*
	 * Checks that the plugin is not in the list of installed
	 * plugins
	 */
	assertThatPluginsIsInstalled(t, testPluginName, false)

	/*
	 * Checks that the plugin is in the list of available
	 * plugins.
	 */
	assertThatPluginsIsAvailable(t, testPluginName, true)
}
