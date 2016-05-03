package registry_test

import (
	"log"
	"os"
	"regexp"
	"runtime"
	"testing"

	"strings"

	"time"

	pb "github.com/eogile/agilestack-core/proto"
	"github.com/eogile/agilestack-core/registry"
	"github.com/eogile/agilestack-utils/dockerclient"
	"github.com/fsouza/go-dockerclient"
)

var dockerClient *dockerclient.DockerClient

/*
 * A Nats server is started before executing the first test
 * and stops after the last one.
 *
 */
var localhostNatsServerURL string

/*
 * Configures the date and time format for logs messages.
 */
func init() {
	log.Println("main init")
	log.SetFlags(log.Lmicroseconds | log.Ldate | log.Ltime)
}

func TestMain(m *testing.M) {
	log.Println("Launching tests")

	/*
	 * Creates the registry
	 */
	dockerWrapper = registry.NewDockerStorageClient()
	testRegistry = registry.NewInMemoryRegistry(dockerWrapper)

	/*
	 * Docker client for test utilities
	 */
	dockerClient = dockerclient.NewClient()

	/*
	 * Cleaning the docker environment before launching the tests
	 */
	uninstallAllPlugins(true)

	/*
	 * Starts the NATS server
	 */
	installNatsServer()

	/*
	 * Computes the URL of the NATS server
	 */
	localhostNatsServerURL = natsServerURL()

	/*
	 * Launching the tests
	 */
	exitCode := m.Run()

	/*
	 * Removing the installed plugins
	 */
	uninstallAllPlugins(true)

	os.Exit(exitCode)
}

/*
 * Helper method that :
 * - Uninstalls all running plugins
 * - Initializes the map of registered plugins
 */
func setUp() {
	uninstallAllRunningPlugins(false)
}

/*
 * Helper method to remove all running containers.
 */
func uninstallAllRunningPlugins(includingNats bool) {
	uninstallPlugins(includingNats, false)
}

/*
 * Helper method to remove all containers including stopped.
 */
func uninstallAllPlugins(includingNats bool) {
	uninstallPlugins(includingNats, true)
}

/*
 * Helper method to remove containers.
 */
func uninstallPlugins(includingNats bool, all bool) {
	listOps := docker.ListContainersOptions{All: all}
	containers, _ := dockerClient.ListContainers(listOps)

	for _, container := range containers {
		if (includingNats && "nats" == container.Image) || strings.HasPrefix(container.Image, "agilestack-") {
			log.Printf("Removing container %s", container.Names[0])
			dockerClient.StopContainer(container.ID, 10)
			removeOpts := docker.RemoveContainerOptions{ID: container.ID}
			dockerClient.RemoveContainer(removeOpts)
		}
	}
}

/*
 * Checks the boolean representing the presence of the given plugin
 * in the list of available plugins equals the given boolean.
 */
func assertThatPluginsIsAvailable(t *testing.T, pluginName string, expectedResult bool) {
	plugins, err := testRegistry.ListAvailablePlugins()
	if err != nil {
		t.Errorf("Error should not be nil : %v", err)
		return
	}

	if expectedResult != pluginsArrayContains(plugins.Plugins, pluginName) {
		t.Errorf("Plugin's availability shoud be %t", expectedResult)
		return
	}
}

/*
 * Checks the boolean representing the presence of the given plugin
 * in the list of installed plugins equals the given boolean.
 */
func assertThatPluginsIsInstalled(t *testing.T, pluginName string, expectedResult bool) {
	plugins, err := testRegistry.ListInstalledPlugins()
	if err != nil {
		t.Errorf("Error should not be nil : %v", err)
		return
	}

	if expectedResult != pluginsArrayContains(plugins.Plugins, pluginName) {
		t.Error("The plugin is present in the list of installed plugins")
		return
	}
}

/*
 * Starts a container running a NATS server
 */
func installNatsServer() {
	/*
	 * Container configuration
	 */
	containerConfig := docker.Config{
		Image: "nats",
	}
	containerConfig.ExposedPorts = map[docker.Port]struct{}{
		"4222/tcp": {},
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"4222/tcp": []docker.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "4222",
			},
		},
	}

	/*
	 * Host configuration
	 */
	hostConfig := docker.HostConfig{
		PublishAllPorts: true,
		NetworkMode:     "agilestacknet",
		PortBindings:    portBindings,
	}

	containerOptions := docker.CreateContainerOptions{
		Name:       "nats",
		Config:     &containerConfig,
		HostConfig: &hostConfig,
	}

	container, err := dockerClient.CreateContainer(containerOptions)
	if err != nil {
		log.Printf("Error on createContainer : %v", err)
	}
	log.Printf("container for image nats created with ID: %s ", container.ID)
	dockerClient.StartContainer(container.ID, nil)
	time.Sleep(500 * time.Millisecond)
}

func pluginsArrayContains(plugins []*pb.Plugin, pluginName string) bool {
	for _, item := range plugins {
		if item.Name == pluginName {
			return true
		}
	}
	return false
}

/*
 * Computes the URL of the NATS server.
 */
func natsServerURL() string {
	dockerEndpoint := os.Getenv("DOCKER_HOST")
	if dockerEndpoint == "" {
		if runtime.GOOS == "darwin" {
			log.Fatal("DOCKER_HOST should be set on a MacOSX")
		}
		return "nats://localhost:4222"
	}

	regex := regexp.MustCompile("tcp://([^:]+):")
	log.Printf("NATS SERVER = %s", "nats://"+regex.FindStringSubmatch(dockerEndpoint)[1]+":4222")
	return "nats://" + regex.FindStringSubmatch(dockerEndpoint)[1] + ":4222"
}
