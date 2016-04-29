package storage_test

import (
	"testing"

	"github.com/eogile/agilestack-core/registry/storage"
	"github.com/fsouza/go-dockerclient"
)

func TestGetPluginNameForSimpleName(t *testing.T) {
	doTestGetPluginName(t, "agilestack-proxy", "agilestack-proxy")
}

func TestGetPluginNameForSimpleNameWithTag(t *testing.T) {
	doTestGetPluginName(t, "agilestack-proxy:latest", "agilestack-proxy")
}

func TestGetPluginNameWithRepository(t *testing.T) {
	doTestGetPluginName(t, "docker-registry.eogile.com/eogile/agilestack-proxy",
		"agilestack-proxy")
}

func TestGetPluginNameWithRepositoryAndTag(t *testing.T) {
	doTestGetPluginName(t, "docker-registry.eogile.com/eogile/agilestack-proxy:1.0",
		"agilestack-proxy")
}

func TestIsContainerAPluginNotAContainer(t *testing.T) {
	container := docker.APIContainers{
		Names: []string{"my-container", "What a wonderful container"},
	}
	if storage.IsContainerAPlugin(container) {
		t.Error("The container should not be considered as a plugin")
	}
}

func TestIsContainerAPlugin(t *testing.T) {
	doTest := func(container docker.APIContainers) {
		if !storage.IsContainerAPlugin(container) {
			t.Error("The container should be considered as a plugin")
		}
	}

	container := docker.APIContainers{
		Names: []string{"my-container", "agilestack-proxy"},
	}
	doTest(container)

	container = docker.APIContainers{
		Names: []string{"my-container", "/agilestack-proxy"},
	}
	doTest(container)
}

func TestTransformContainers(t *testing.T) {
	containers := []docker.APIContainers{
		docker.APIContainers{
			Names: []string{"container1", "not_a_plugin"},
			Image: "agilestack-proxy",
		},
		docker.APIContainers{
			Names: []string{"container2", "/agilestack-proxy"},
			Image: "docker-registry.eogile.com/eogile/agilestack-proxy:latest",
		},
		docker.APIContainers{
			Names: []string{"container3", "agilestack-back"},
			Image: "agilestack-back",
		},
	}

	plugins := storage.TransformContainers(containers).Plugins

	if len(plugins) != 2 {
		t.Fatalf("Invalid number of plugins (%d)\n", len(plugins))
	}
	if plugins[0].Name != "agilestack-proxy" {
		t.Errorf("Invalid name for the first plugin: %s\n", plugins[0].Name)
	}
	if plugins[1].Name != "agilestack-back" {
		t.Errorf("Invalid name for the second plugin: %s\n", plugins[1].Name)
	}

}

func doTestGetPluginName(t *testing.T, imageName string, expectedName string) {
	pluginName := storage.GetPluginName(imageName)
	if pluginName != expectedName {
		t.Errorf("Invalid plugin name: %s\n", pluginName)
	}
}
