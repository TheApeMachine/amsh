package container

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

type Builder struct {
	client *client.Client
}

func NewBuilder() (*Builder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Builder{client: cli}, nil
}

func (b *Builder) BuildImage(ctx context.Context, dockerfilePath, imageName string) error {
	tar, err := archive.TarWithOptions(dockerfilePath, &archive.TarOptions{})
	if err != nil {
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageName},
		Remove:     true,
	}

	resp, err := b.client.ImageBuild(ctx, tar, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}
