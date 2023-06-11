package utils

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Container struct {
	testcontainers.Container
	HostIP string
	Port   int
}

type Image struct {
	Name string
	Port string
}

var (
	MongoDB = Image{
		Name: "mongo",
		Port: "27017",
	}
	RabbitMQ = Image{
		Name: "rabbitmq",
		Port: "5672",
	}
	Images = []Image{MongoDB, RabbitMQ}
)

func PrepareContainer(ctx context.Context, image Image, version ...string) (*Container, error) {
	dockerImage := image.Name
	if len(version) > 0 && version[0] != "" {
		dockerImage += ":" + version[0]
	}
	req := testcontainers.ContainerRequest{
		Name:         image.Name + "TestContainer",
		Image:        dockerImage,
		ExposedPorts: []string{image.Port + "/tcp"},
		WaitingFor:   wait.ForListeningPort(nat.Port(image.Port)),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Reuse:            true,
		Started:          true,
		Logger:           &log.Logger,
	})
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(image.Port))
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("TestContainers: container %s is now running at %s:%d", req.Image, hostIP, mappedPort.Int())
	return &Container{Container: container, HostIP: hostIP, Port: mappedPort.Int()}, nil
}
