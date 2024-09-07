package container

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Runner struct {
	client *client.Client
}

func NewRunner() (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Runner{client: cli}, nil
}

func (r *Runner) RunContainer(ctx context.Context, imageName string, cmd []string) error {
	resp, err := r.client.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   cmd,
		Tty:   true,
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	if err := r.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := r.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	out, err := r.client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, out)
	return err
}

func (r *Runner) StopContainer(ctx context.Context, containerID string) error {
	timeout := 10 // seconds
	return r.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}