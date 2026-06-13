package main

import (
	"context"
	"fmt"
	"os"
	"net"
	"os/user"
	"runtime"

	nettypes "go.podman.io/common/libnetwork/types"
	"go.podman.io/podman/v6/libpod/define"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/containers"
	"go.podman.io/podman/v6/pkg/specgen"
)

func GetPlatformSocket() (*string, error) {
	platform := runtime.GOOS
	if platform == "windows" {
		pipeDir := "npipe:////./pipe/podman-machine-default"
		return &pipeDir, nil
	} else if platform == "darwin" {
		user, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("Failed to get the current user: %w", err)
		}
		socketDir := fmt.Sprintf("unix://%s/.local/share/containers/podman/machine/podman.sock", user.HomeDir)
		return &socketDir, nil
	} else {
		user, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("Failed to get the current user: %w", err)
		}
		socketDir := fmt.Sprintf("unix:///run/user/%s/podman/podman.sock", user.Uid)
		return &socketDir, nil
	}
}

type Connection struct {
	provisionCount int
	client *context.Context
}

func NewPodman() (*Connection, error) {
	platformSocket, err := GetPlatformSocket()
	if err != nil {
		return nil, fmt.Errorf("Failed to get socket: %w", err)
	}

	client, err := bindings.NewConnection(context.Background(), *platformSocket)
	if err != nil {
		return nil, err
	}
	return &Connection{provisionCount: 0, client: &client}, nil
}

func (client *Connection) Inspect(container string) (*define.InspectContainerData, error) {
	inspectData, err := containers.Inspect(
		*client.client,
		container,
		new(containers.InspectOptions).WithSize(true),
	)
	if err != nil {
		return nil, err
	}
	return inspectData, nil
}

func (client *Connection) CreateContainer(image string, name string, ip string) error {
	// Generate the spec for the container
	spec := specgen.NewSpecGenerator(image, false)
	spec.Name = name
	spec.Systemd = "always"
	// Boolean members require a pointer to a boolean, so we need it to
	// be assigned as such *sigh*
	privileged := true
	spec.Remove = &privileged
	spec.Privileged = &privileged

	_, err := containers.CreateWithSpec(*client.client, spec, nil)
	if err != nil {
		return err
	}

	client.provisionCount++
	return nil
}
func (client *Connection) StartContainer(name string) error {
	if err := containers.Start(*client.client, name, nil); err != nil {
		return err
	}
	return nil
}

func (client *Connection) StopContainer(name string) error {
	if err := containers.Stop(*client.client, name, nil); err != nil {
		return err
	}
	return nil
}

func (client *Connection) DeleteContainer(name string) error {
	if _, err := containers.Remove(*client.client, name, nil); err != nil {
		return err
	}

	client.provisionCount--
	return nil
}

func (client *Connection) ListContainers() ([]string, error) {
	containers, err := containers.List(*client.client, nil)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, container := range containers {
		names = append(names, container.Names[0])
	}
	return names, nil
}
