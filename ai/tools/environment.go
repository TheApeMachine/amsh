package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/container"
	"github.com/theapemachine/errnie"
)

type Environment struct {
	Command string `json:"command"`
	runner  *container.Runner
}

func NewEnvironment() *Environment {
	runner, err := container.NewRunner()

	if err != nil {
		errnie.Error(err)
		return nil
	}

	return &Environment{runner: runner}
}

func (environment *Environment) GenerateSchema() string {
	schema := jsonschema.Reflect(&Environment{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (environment *Environment) Use(ctx context.Context, args map[string]any) string {
	result, err := environment.Execute(ctx, args)
	if err != nil {
		errnie.Error(err)
	}
	return result
}

// Execute initializes a container environment and runs commands inside it
func (environment *Environment) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Retrieve container name from arguments
	name, err := getStringArg(args, "name", "")
	if err != nil || name == "" {
		return "", errors.New("container name is required")
	}

	// Optional: retrieve command from args, default to interactive shell
	cmd := []string{"/bin/sh"}
	if customCmd, ok := args["cmd"].([]string); ok && len(customCmd) > 0 {
		cmd = customCmd
	}

	// Build container image based on given name
	builder, err := container.NewBuilder()
	if err != nil {
		return "", fmt.Errorf("failed to create container builder: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	imagePath := filepath.Join(wd, "container")
	if err := builder.BuildImage(ctx, imagePath, name); err != nil {
		return "", fmt.Errorf("failed to build container image: %w", err)
	}

	// Run container and attach to it
	stdin, stdout, err := environment.runner.RunContainer(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}
	defer stdin.Close()

	// Execute the command in the container
	output, err := environment.runCommandInContainer(stdin, stdout, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute command in container: %w", err)
	}

	return output, nil
}

// runCommandInContainer runs a command in the container and reads the output
func (environment *Environment) runCommandInContainer(stdin io.WriteCloser, stdout io.ReadCloser, cmd []string) (string, error) {
	command := []byte(fmt.Sprintf("%s\n", cmd))
	if _, err := stdin.Write(command); err != nil {
		return "", fmt.Errorf("failed to send command to container: %w", err)
	}

	// Read command output
	var result strings.Builder
	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			result.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read from container stdout: %w", err)
		}
	}

	return result.String(), nil
}
