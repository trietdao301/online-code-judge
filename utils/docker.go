package utils

import (
	"log"

	"github.com/docker/docker/client"
)

func InitializeDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error initializing Docker client: %v", err)
		return nil, err
	}
	return cli, nil
}
