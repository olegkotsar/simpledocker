package simpledocker

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type SimpleContainer struct {
	client    *client.Client
	container string
}

// publish - "2222:22", "127.0.0.1:80:8080"
func CreateContainer(
	name string,
	image string,
	publish []string,
	env map[string]string,
) (*SimpleContainer, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	if err = checkImageExists(c, image); err != nil {
		return nil, err
	}

	// Create host config.
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	if len(publish) != 0 {
		_, portsBindings, err := nat.ParsePortSpecs(publish)
		if err != nil {
			return nil, err
		}
		hostConfig.PortBindings = portsBindings
	}

	// Create network config.
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	networkConfig := &network.NetworkingConfig{}

	// Create container config.
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		Image: image,
	}

	if len(env) != 0 {
		dockerenv := make([]string, 0, len(env))
		for k, v := range env {
			dockerenv = append(dockerenv, k+"="+v)
		}
		config.Env = dockerenv
	}

	ctx := context.Background()
	// Create container.
	cont, err := c.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkConfig,
		nil,
		name,
	)
	if err != nil {
		return nil, err
	}

	// Start container.
	err = c.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}
	return &SimpleContainer{
		client:    c,
		container: cont.ID,
	}, nil
}

func (sc *SimpleContainer) Stop() error {
	ctx := context.Background()
	return sc.client.ContainerStop(ctx, sc.container, nil)
}

func (sc *SimpleContainer) Remove() error {
	ctx := context.Background()

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	return sc.client.ContainerRemove(ctx, sc.container, removeOptions)
}

func checkImageExists(c *client.Client, image string) error {
	ctx := context.Background()
	list, err := c.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}
	imageExists := false
	for _, v := range list {
		for _, tag := range v.RepoTags {
			if tag == image {
				imageExists = true
				break
			}
		}
	}
	if !imageExists {
		reader, err := c.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()
		io.Copy(os.Stdout, reader)
	}
	return nil
}
