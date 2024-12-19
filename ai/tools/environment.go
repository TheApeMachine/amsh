package tools

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/theapemachine/amsh/container"
	"github.com/theapemachine/errnie"
)

type Environment struct {
	Task string `json:"task"`
	IO   io.ReadWriteCloser
}

func NewEnvironment() *Environment {
	return &Environment{}
}

func (environment *Environment) GenerateSchema() string {
	return ""
}

func (environment *Environment) Use(ctx context.Context, args map[string]any) string {
	builder := errnie.SafeMust(func() (*container.Builder, error) {
		return container.NewBuilder()
	})

	wd := errnie.SafeMust(func() (string, error) {
		return os.Getwd()
	})

	if errnie.Error(builder.BuildImage(context.Background(), filepath.Join(wd, "container", "Dockerfile"), "test")) != nil {
		return "failed to build image"
	}

	runner := errnie.SafeMust(func() (*container.Runner, error) {
		return container.NewRunner()
	})

	// Create directories if they don't exist
	os.MkdirAll("/tmp/out", 0755)
	os.MkdirAll("/tmp/.ssh", 0755)

	conn, err := runner.RunContainer(ctx, "test")
	if err != nil {
		errnie.Error(err)
		return "failed to start container"
	}
	environment.IO = conn

	// Send an initial command to get the prompt
	if _, err := conn.Write([]byte("echo '" + args["task"].(string) + "'\n")); err != nil {
		errnie.Error(err)
		return "failed to initialize connection"
	}

	return "container ready"
}

func (environment *Environment) GetIO() io.ReadWriteCloser {
	return environment.IO
}

func (environment *Environment) IsInteractive() bool {
	return true
}
