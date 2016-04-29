package registry

import (
	"os"

	pb "github.com/eogile/agilestack-core/proto"
	"encoding/json"
	"github.com/fsouza/go-dockerclient"
	"log"
)

const (
	PLUGIN_TEMPLATE_DIR = "/plugin-template"
)

type (
	pluginFactory interface {
		CreatePlugin(request *pb.NewPluginRequest) error
	}

	dockerPluginFactory struct {
		/*
		 * The client to access the location where the plugins are installed.
		 */
		dockerClient *DockerStorageClient
	}

	configuration struct {
		Url  string `json:"url"`
		Name string `json:"name"`
	}
)

func NewPluginFactory() pluginFactory {
	return &dockerPluginFactory{
		dockerClient: NewDockerStorageClient(),
	}
}

/*
 * 1 - Go to the template directory
 * 2 - Compile the JavaScript code
 * 3 - Create the Docker image
 */
func (factory dockerPluginFactory) CreatePlugin(request *pb.NewPluginRequest) error {
	/*
	 * Creating the configuration file.
	 */
	err := factory.createConfigurationFile(request)
	if err != nil {
		return err
	}

	/*
	 * Building the Docker image.
	 */
	options := docker.BuildImageOptions{
		Name:         "agilestack-" + request.Name,
		ContextDir:   request.Directory,
		OutputStream: os.Stdout,
	}
	return factory.dockerClient.docker.BuildImage(options)
}

func (factory dockerPluginFactory) createConfigurationFile(request *pb.NewPluginRequest) error {
	file, err := os.Create(request.Directory + "/config.json")
	if err != nil {
		log.Println("Error while creating configuration file", err)
		return err
	}
	defer file.Close()

	config := &configuration{
		Name: request.Name,
		Url:  request.Url,
	}
	return json.NewEncoder(file).Encode(config)
}
