package container

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

/*
Runner encapsulates the functionality for running and interacting with Docker containers.
It provides methods to create, start, and attach to containers, allowing for seamless
integration with external systems such as language models.
*/
type Runner struct {
	client      *client.Client
	containerID string
	in          io.WriteCloser
	out         io.ReadCloser
}

/*
NewRunner initializes a new Runner instance with a Docker client.
This setup allows for interaction with the Docker daemon on the host system.
*/
func NewRunner() (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &Runner{client: cli}, nil
}

/*
RunContainer creates, starts, and attaches to a new container based on the specified image.
It provides channels for stdin and stdout/stderr, enabling interactive communication with the container.
This method is particularly useful for integrating with language models or other interactive processes.

Parameters:
  - ctx: The context for the Docker API calls
  - imageName: The name of the Docker image to use
  - cmd: The command to run in the container
  - username: The username to create and use within the container
  - customMessage: A message to be displayed when attaching to the container

Returns:
  - in: A channel for sending input to the container
  - out: A channel for receiving output from the container
  - err: Any error encountered during the process
*/
func (r *Runner) RunContainer(ctx context.Context, imageName string, cmd []string, username, customMessage string) (in io.WriteCloser, out io.ReadCloser, err error) {
	// Create host config with volume mount
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/tmp/workspace",
				Target: "/workspace",
			},
		},
	}

	// Create the container with specific configuration
	resp, err := r.client.ContainerCreate(ctx, &container.Config{
		Image:     imageName,
		Cmd:       cmd,
		Tty:       true,
		OpenStdin: true,
		StdinOnce: false,
		Env: []string{
			fmt.Sprintf("USERNAME=%s", username),
			fmt.Sprintf("CUSTOM_MESSAGE=%s", customMessage),
		},
		WorkingDir: "/workspace", // Set the working directory to the mounted volume
	}, hostConfig, nil, nil, "")
	if err != nil {
		return nil, nil, err
	}

	// Start the container
	if err := r.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, nil, err
	}

	// Attach to the container
	attachResp, err := r.client.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})

	if err != nil {
		return nil, nil, err
	}

	// Create a multiplexed reader for stdout and stderr
	r.out, r.in = io.Pipe()
	go func() {
		_, _ = stdcopy.StdCopy(r.in, r.in, attachResp.Reader)
	}()

	fmt.Printf("Container %s is running with user %s\n", resp.ID, username)

	r.containerID = resp.ID

	return attachResp.Conn, r.out, nil
}

/*
StopContainer gracefully stops a running container.
This method ensures proper cleanup of container resources.

Parameters:
  - ctx: The context for the Docker API calls
  - containerID: The ID of the container to stop
*/
func (r *Runner) StopContainer(ctx context.Context) error {
	timeout := 10 // seconds
	return r.client.ContainerStop(ctx, r.containerID, container.StopOptions{Timeout: &timeout})
}

/*
ExecuteCommand executes a command in the container and returns the output.
*/
func (r *Runner) ExecuteCommand(ctx context.Context, cmd []string) ([]byte, error) {
	r.in.Write([]byte(strings.Join(cmd, " ")))
	return io.ReadAll(r.out)
}
