package storage

import (
	"log"
	"strings"

	pb "github.com/eogile/agilestack-core/proto"
	"github.com/fsouza/go-dockerclient"
)

type DockerHelper struct {
	docker *docker.Client
}

func NewDockerHelper(docker *docker.Client) *DockerHelper {
	return &DockerHelper{
		docker: docker,
	}
}

/*
 * Returns a boolean indicating whether or not the given image is a
 * AgileStack plugin or not.
 *
 * A Docker image is considered as a plugin if it is referenced by
 * at least one repository whose name (without tag, registry and user)
 * starts with "agilestack-".
 */
func (h *DockerHelper) IsImageAPlugin(image docker.APIImages) bool {
	return strings.HasPrefix(
		h.GetPluginName(image), "agilestack-")
}

/*
 * Extracts the name of the plugin matching the given Docker image.
 *
 * The plugin's name is based on the name of the Docker image.
 * The image's tag, the registry (if present) and the image owner are skipped.
 *
 * Please notice that in the case where the image is referenced by several
 * repositories, the first one is used to deduce the plugin's name.
 *
 * Examples :
 * - docker-registry.eogile.com/eogile/agilestack-proxy:latest => agilestack-proxy
 * - agilestack-proxy:latest => agilestack-proxy
 */
func (h *DockerHelper) GetPluginName(image docker.APIImages) string {
	return GetPluginName(image.RepoTags[0])
}

/*
 * Returns the Docker image from which the given plugin was created.
 *
 * The match between an image and a plugin is computed on the image's
 * name.
 */
func (h *DockerHelper) ImageFromPlugin(pluginName string) *docker.APIImages {

	/*
	 * Listing the images
	 */
	images, imageErr := h.ListImages()
	if imageErr != nil {
		log.Printf("Error when listing Docker images : %v", imageErr)
		return nil
	}

	for _, image := range images {
		if h.GetPluginName(image) == pluginName {
			return &image
		}
	}
	return nil
}

/*
 * Returns the list of top-level Docker images.
 */
func (h *DockerHelper) ListImages() ([]docker.APIImages, error) {
	listOps := docker.ListImagesOptions{All: false}
	return h.docker.ListImages(listOps)
}

/*
 * Transforms the given containers into plugins objects.
 *
 * Only containers that are plugins are transformed.
 */
func TransformContainers(containers []docker.APIContainers) *pb.Plugins {
	plugins := &pb.Plugins{Plugins: make([]*pb.Plugin, 0)}
	for _, container := range containers {

		if IsContainerAPlugin(container) {

			pluginName := GetPluginName(container.Image)
			plugin := &pb.Plugin{Name: pluginName, PluginStatus: pb.PluginStatus_OK}
			for _, APIPort := range container.Ports {
				log.Printf("containerImage : %v, privatePort : %v, publicPort : %v",
					pluginName, APIPort.PrivatePort, APIPort.PublicPort)
			}
			plugins.Plugins = append(plugins.Plugins, plugin)

		}
	}
	return plugins
}

/*
 * Extracts the name of the plugin matching the given Docker image's name.
 *
 * The plugin's name is based on the name of the Docker image.
 * The image's tag, the registry (if present) and the image owner are skipped.
 *
 * Please notice that in the case where the image is referenced by several
 * repositories, the first one is used to deduce the plugin's name.
 *
 * Examples :
 * - docker-registry.eogile.com/eogile/agilestack-proxy:latest => agilestack-proxy
 * - agilestack-proxy:latest => agilestack-proxy
 */
func GetPluginName(imageName string) string {
	if strings.Contains(imageName, "/") {
		items := strings.Split(imageName, "/")
		imageName = items[len(items)-1]
	}

	if strings.Contains(imageName, ":") {
		items := strings.Split(imageName, ":")
		imageName = items[0]
	}

	return imageName
}

/*
 * Returns a boolean indicating whether or not the given container is a
 * AgileStack plugin or not.
 *
 * A Docker container is considered as a plugin if it's name starts with
 * "agilestack-".
 *
 * TODO The correct way would be to find the image and to call IsImageAPlugin...
 */
func IsContainerAPlugin(container docker.APIContainers) bool {

	for _, containerName := range container.Names {
		name := containerName
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}
		if strings.HasPrefix(name, "agilestack-") {
			return true
		}
	}
	return false
}
