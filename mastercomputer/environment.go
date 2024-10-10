package mastercomputer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Environment struct {
	Function     *openai.FunctionDefinition
	docker       *client.Client
	containerID  string
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	outputBuffer *bytes.Buffer
	mu           sync.Mutex
}

func NewEnvironment() *Environment {
	errnie.Trace()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		errnie.Error(err)
		return nil
	}

	return &Environment{
		Function: &openai.FunctionDefinition{
			Name:        "environment",
			Description: "Use a fully functional Linux environment. You will be directly connected to stdin and stdout, so from the moment you use this, you will need to only respond with valid commands until you exit.",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Properties: map[string]jsonschema.Definition{
					"shell": {
						Type:        jsonschema.String,
						Description: "The shell you want to use.",
						Enum:        []string{"bash", "zsh", "sh"},
					},
				},
				Required: []string{"shell"},
			},
		},
		docker:       cli,
		outputBuffer: &bytes.Buffer{},
	}
}

func (env *Environment) Start(ctx context.Context, shell string) error {
	errnie.Trace()

	// Create and start the container
	resp, err := env.docker.ContainerCreate(ctx, &container.Config{
		Image:     "ubuntu:latest",
		Cmd:       []string{shell},
		Tty:       true,
		OpenStdin: true,
	}, nil, nil, nil, "")
	if errnie.Error(err) != nil {
		return err
	}

	env.containerID = resp.ID

	if err := env.docker.ContainerStart(ctx, env.containerID, container.StartOptions{}); err != nil {
		return errnie.Error(err)
	}

	waiter, err := env.docker.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		return errnie.Error(err)
	}

	env.stdin = waiter.Conn
	go env.handleOutput(ctx, waiter.Reader)

	return nil
}

func (env *Environment) handleOutput(ctx context.Context, reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Container output:", line)
		env.mu.Lock()
		env.outputBuffer.WriteString(line + "\n")
		env.mu.Unlock()
	}
}

// WriteToStdin writes input to the container's stdin.
func (env *Environment) WriteToStdin(input string) error {
	_, err := fmt.Fprintln(env.stdin, input)
	return err
}

// ReadOutputBuffer reads output accumulated in the buffer.
func (env *Environment) ReadOutputBuffer() string {
	env.mu.Lock()
	defer env.mu.Unlock()
	output := env.outputBuffer.String()
	env.outputBuffer.Reset()
	return output
}

// Close gracefully stops and removes the container
func (env *Environment) Close() error {
	if env.stdin != nil {
		env.stdin.Close()
	}
	if env.containerID != "" {
		return env.docker.ContainerRemove(context.Background(), env.containerID, container.RemoveOptions{Force: true})
	}
	return nil
}
