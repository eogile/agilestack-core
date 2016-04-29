package registry

import (
	"log"
	"strings"

	"os"

	pb "github.com/eogile/agilestack-core/proto"
	"github.com/eogile/agilestack-core/registry/storage"
	"github.com/eogile/agilestack-utils/dockerclient"
	"github.com/fsouza/go-dockerclient"
)

type PluginStorageClient interface {
	ListInstallablePlugins() (*pb.Plugins, error)

	ListInstalledPlugins() (*pb.Plugins, error)

	InstallPlugin(imageName string, cmd string) error

	UninstallPlugin(imageName string) error

	/*
	 * Returns a boolean indicating whether or not the given
	 * plugin is currently installed.
	 *
	 * If a plugin is installed but not running, "true" will
	 * be returned.
	 *
	 * If the returned error is not "nil", then the boolean
	 * value is irrelevant.
	 */
	IsPluginInstalled(name string) (bool, error)
}

type DockerStorageClient struct {
	docker    *docker.Client
	sharedDir string
	helper    *storage.DockerHelper
}

func NewDockerStorageClient() *DockerStorageClient {
	sharedDir := os.Getenv("SHARED_FOLDER")
	if sharedDir == "" {
		log.Fatal("$SHARED_FOLDER is undefined. Cannot initialize the plugins.")
	}
	log.Println("Shared folder :", sharedDir)
	docker := dockerclient.NewClient().Client

	return &DockerStorageClient{
		docker:    docker,
		sharedDir: sharedDir,
		helper:    storage.NewDockerHelper(docker),
	}
}

func (dockerWrapper *DockerStorageClient) ListInstallablePlugins() (*pb.Plugins, error) {
	/*
	 * Listing containers
	 */
	containers, err := dockerWrapper.listRunningContainers()
	if err != nil {
		log.Printf("Error when listing running Docker containers : %v", err)
		return nil, err
	}

	/*
	 * Listing images
	 */
	images, imageErr := dockerWrapper.helper.ListImages()
	if imageErr != nil {
		log.Printf("Error when listing Docker images : %v", imageErr)
		return nil, imageErr
	}

	/*
	 * Filters the list of images to exclude the images of
	 * the running containers.
	 */
	runningPlugins := storage.TransformContainers(containers).Plugins
	plugins := &pb.Plugins{Plugins: make([]*pb.Plugin, 0)}
	for _, image := range images {
		if dockerWrapper.helper.IsImageAPlugin(image) {
			pluginName := dockerWrapper.helper.GetPluginName(image)

			if !pluginsArrayContains(runningPlugins, pluginName) {
				plugin := &pb.Plugin{Name: pluginName}
				plugins.Plugins = append(plugins.Plugins, plugin)
			}
		}
	}
	return plugins, nil
}

func (dockerWrapper *DockerStorageClient) ListInstalledPlugins() (*pb.Plugins, error) {
	/*
	 * Listing containers
	 */
	containers, err := dockerWrapper.listRunningContainers()
	if err != nil {
		log.Printf("Error when listing running Docker containers : %v", err)
		return nil, err
	}
	return storage.TransformContainers(containers), nil
}

func (dockerWrapper *DockerStorageClient) InstallPlugin(pluginName string, cmd string) error {
	log.Printf("Creating container for plugin %s", pluginName)

	/*
	 * Finding the Docker image matching the plugin's name
	 */
	image := dockerWrapper.helper.ImageFromPlugin(pluginName)
	log.Printf("Creating container for image %s with cmd %s", image.RepoTags[0], cmd)

	/*
	 * Container configuration
	 */
	containerConfig := docker.Config{
		Image: image.RepoTags[0],
	}
	if len(cmd) > 0 {
		containerConfig.Cmd = append(containerConfig.Cmd, cmd)
	}

	/*
	 * Host configuration
	 */
	hostConfig := docker.HostConfig{
		PublishAllPorts: true,
		Binds:           []string{"agilestack-shared:/shared"},
		NetworkMode:     "agilestacknet",
	}

	containerOptions := docker.CreateContainerOptions{
		Name:       pluginName,
		Config:     &containerConfig,
		HostConfig: &hostConfig,
	}

	container, err := dockerWrapper.docker.CreateContainer(containerOptions)
	if err != nil {
		log.Printf("Error on createContainer : %v", err)
		return err
	}
	log.Printf("container created with ID: %s ", container.ID)

	err = dockerWrapper.docker.StartContainer(container.ID, nil)
	if err != nil {
		log.Printf("Error on startContainer : %v", err)
		return err
	}
	attachContainerOptions := docker.AttachToContainerOptions{
		Container: image.RepoTags[0],
	}
	dockerWrapper.docker.AttachToContainer(attachContainerOptions)
	return nil
}

func (dockerWrapper *DockerStorageClient) UninstallPlugin(pluginName string) error {
	/*
	 * Listing all the containers, even the stopped ones.
	 */
	listOps := docker.ListContainersOptions{All: true}
	containers, err := dockerWrapper.docker.ListContainers(listOps)
	if err != nil {
		log.Printf("Error when listing running Docker containers : %v", err)
		return err
	}
	for _, container := range containers {
		for _, containerName := range container.Names {

			if containerName == pluginName || containerName == "/"+pluginName {
				log.Printf("container %s status : %s", container.ID, container.Status)
				if strings.HasPrefix(container.Status, "Up") {
					/*
					 * Stopping the container with timeout
					 */
					dockerWrapper.docker.StopContainer(container.ID, 10)
				}

				removeOpts := docker.RemoveContainerOptions{
					ID: container.ID}
				dockerWrapper.docker.RemoveContainer(removeOpts)
				log.Printf("container %s removed", container.ID)

			}
		}
	}
	return nil
}

func (dockerWrapper *DockerStorageClient) IsPluginInstalled(name string) (bool, error) {
	/*
	 * Listing all the containers, even the stopped ones.
	 */
	listOps := docker.ListContainersOptions{All: true}
	containers, err := dockerWrapper.docker.ListContainers(listOps)

	if err != nil {
		return false, err
	}

	plugins := storage.TransformContainers(containers)
	return pluginsArrayContains(plugins.Plugins, name), nil
}

func (dockerWrapper *DockerStorageClient) listRunningContainers() ([]docker.APIContainers, error) {
	listOps := docker.ListContainersOptions{All: false}
	return dockerWrapper.docker.ListContainers(listOps)
}

func pluginsArrayContains(plugins []*pb.Plugin, pluginName string) bool {
	for _, item := range plugins {
		if item.Name == pluginName {
			return true
		}
	}
	return false
}
