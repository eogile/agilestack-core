package main

import (
	"log"
	"os"
	"testing"

	"github.com/eogile/agilestack-utils/dockerclient"
	"github.com/eogile/agilestack-utils/slices"
	"github.com/fsouza/go-dockerclient"
)

var dockerClient *dockerclient.DockerClient

const network = "agilestacknet"

func TestMain(m *testing.M) {
	log.Println("Launching tests agileStack")

	/*
	 * Docker client for test utilities
	 */
	dockerClient = dockerclient.NewClient()

	/*
	 * Create agilestacknet docker network if not exists
	 */
	networks, err := dockerClient.ListNetworks()
	if err != nil {
		log.Println("unable to List docker networks : ", err)
		os.Exit(1)
	}

	options := docker.CreateNetworkOptions{
		Name:   network,
		Driver: "bridge",
	}
	if !slices.DockerNetworkInSlice(network, networks) {
		//create network
		_, errNet := dockerClient.CreateNetwork(options)
		if errNet != nil {
			log.Printf("Cannot create docker network %v. Got error %v", network, errNet)
		}
	}

	os.Exit(0)
}
