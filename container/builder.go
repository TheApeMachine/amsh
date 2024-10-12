package container

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

/*
Builder is a wrapper around the Docker client that provides methods for building
and running containers. It encapsulates the complexity of Docker operations,
allowing for easier management of containerized environments.
*/
type Builder struct {
	client *client.Client
}

/*
NewBuilder creates a new Builder instance.
It initializes a Docker client using the host's Docker environment settings.
*/
func NewBuilder() (*Builder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Builder{client: cli}, nil
}

/*
BuildImage constructs a Docker image from a Dockerfile in the specified directory.
This method abstracts the image building process, handling the creation of a tar archive
and the configuration of build options.
*/
func (b *Builder) BuildImage(ctx context.Context, dockerfilePath, imageName string) error {
	tar, err := archive.TarWithOptions(dockerfilePath, &archive.TarOptions{})
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageName},
		Remove:     true,
		// Add these options for better compatibility:
		BuildArgs: map[string]*string{
			"TARGETARCH": nil, // This will use the default architecture
		},
		Target: "dev", // Build the dev stage by default
	}

	resp, err := b.client.ImageBuild(ctx, tar, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}
