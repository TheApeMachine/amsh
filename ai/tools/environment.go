package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/container"
)

type EnvironmentTool struct {
	runner *container.Runner
}

func NewEnvironmentTool() (*EnvironmentTool, error) {
	runner, err := container.NewRunner()
	if err != nil {
		return nil, err
	}
	return &EnvironmentTool{runner: runner}, nil
}

// Description retrieves the description from the Viper configuration
func (e *EnvironmentTool) Description() string {
	return viper.GetViper().GetString("tools.environment")
}

// Execute initializes a container environment and runs commands inside it
func (e *EnvironmentTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
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
	stdin, stdout, err := e.runner.RunContainer(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}
	defer stdin.Close()

	// Execute the command in the container
	output, err := e.runCommandInContainer(ctx, stdin, stdout, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute command in container: %w", err)
	}

	return output, nil
}

// runCommandInContainer runs a command in the container and reads the output
func (e *EnvironmentTool) runCommandInContainer(ctx context.Context, stdin io.WriteCloser, stdout io.ReadCloser, cmd []string) (string, error) {
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
