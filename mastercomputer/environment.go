package mastercomputer

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type EnvironmentJob struct {
	ctx          context.Context
	shell        string
	Function     *openai.FunctionDefinition
	docker       *client.Client
	containerID  string
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stdinChan    chan string
	outputBuffer *bytes.Buffer
	mu           sync.Mutex
}

func NewEnvironmentJob(ctx context.Context, shell string) *EnvironmentJob {
	errnie.Trace()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		errnie.Error(err)
		return nil
	}

	return &EnvironmentJob{
		ctx:          ctx,
		shell:        shell,
		docker:       cli,
		outputBuffer: &bytes.Buffer{},
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
	}
}

// Implement the Job interface
func (ej *EnvironmentJob) Process(ctx context.Context) error {
	errnie.Trace()
	return ej.Start()
}

func (ej *EnvironmentJob) Start() error {
	errnie.Trace()

	// Create and start the container
	resp, err := ej.docker.ContainerCreate(ej.ctx, &container.Config{
		Image:     "ubuntu:latest",
		Cmd:       []string{ej.shell},
		Tty:       true,
		OpenStdin: true,
	}, nil, nil, nil, "")
	if errnie.Error(err) != nil {
		return err
	}

	ej.containerID = resp.ID

	if err := ej.docker.ContainerStart(ej.ctx, ej.containerID, container.StartOptions{}); err != nil {
		return errnie.Error(err)
	}

	waiter, err := ej.docker.ContainerAttach(ej.ctx, resp.ID, container.AttachOptions{
		Stderr: true,
		Stdout: true,
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		return errnie.Error(err)
	}

	ej.stdin = waiter.Conn
	go ej.handleInput(ej.ctx)
	go ej.handleOutput(ej.ctx, waiter.Reader)

	// Keep the environment running
	<-ej.ctx.Done()
	return ej.Close()
}

func (ej *EnvironmentJob) handleInput(ctx context.Context) {
	for {
		select {
		case input, ok := <-ej.stdinChan:
			if !ok {
				return
			}
			_, err := fmt.Fprintln(ej.stdin, input)
			if err != nil {
				errnie.Error(err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ej *EnvironmentJob) handleOutput(ctx context.Context, reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Container output:", line)
		ej.mu.Lock()
		ej.outputBuffer.WriteString(line + "\n")
		ej.mu.Unlock()
	}
}

func (ej *EnvironmentJob) WriteToStdin(input string) error {
	select {
	case ej.stdinChan <- input:
		return nil
	default:
		return errnie.Error(errors.New("stdin channel is full"))
	}
}

// ReadOutputBuffer reads output accumulated in the buffer.
func (ej *EnvironmentJob) ReadOutputBuffer() string {
	ej.mu.Lock()
	defer ej.mu.Unlock()
	output := ej.outputBuffer.String()
	ej.outputBuffer.Reset()
	return output
}

// Close gracefully stops and removes the container
func (ej *EnvironmentJob) Close() error {
	if ej.stdin != nil {
		ej.stdin.Close()
	}
	if ej.containerID != "" {
		return ej.docker.ContainerStop(context.Background(), ej.containerID, container.StopOptions{})
	}
	return nil
}
