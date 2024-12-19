package marvin

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
)

type ToolHandler struct {
	agent *Agent
	inout io.ReadWriteCloser
}

func NewToolHandler(agent *Agent) *ToolHandler {
	return &ToolHandler{
		agent: agent,
	}
}

func (toolHandler *ToolHandler) Initialize() {
	_ = toolHandler.agent.tool.Use(
		context.Background(),
		map[string]any{"task": "system ready"},
	)

	toolHandler.inout = toolHandler.agent.tool.(ai.InteractiveTool).GetIO()

	if toolHandler.inout == nil {
		errnie.Error(fmt.Errorf("tool IO not available"))
	}
}

// handleInteractiveTool manages the IO stream for interactive tools
func (toolHandler *ToolHandler) Accumulator() *twoface.Accumulator {
	return twoface.NewAccumulator(
		toolHandler.agent.Name,
		"assistant",
		"command",
	).Yield(func(acc *twoface.Accumulator) {
		defer close(acc.Out)

		// Generate next command
		for artifact := range toolHandler.agent.provider.Generate(toolHandler.agent.buffer.Peek()) {
			acc.Out <- artifact
		}

		// Execute command and handle response
		command := acc.Take().Peek("payload")
		if err := toolHandler.executeCommand(command, acc.Out); err != nil {
			errnie.Error(err)
			return
		}

		// Update message history with both command and response
		toolHandler.agent.buffer.Poke(data.New(
			toolHandler.agent.Name,
			"assistant",
			"command",
			[]byte(command),
		))
	})
}

// Helper function to execute commands and handle responses
func (toolHandler *ToolHandler) executeCommand(command string, out chan<- *data.Artifact) error {
	// Write command
	if _, err := toolHandler.inout.Write([]byte(command + "\n")); err != nil {
		return err
	}

	buffer := make([]byte, 4096)
	promptEnd := []byte("# ")

	for {
		n, err := toolHandler.inout.Read(buffer)
		if err != nil {
			return err
		}

		if n > 0 {
			chunk := buffer[:n]
			out <- data.New(toolHandler.agent.Name, "assistant", "tool", chunk)

			if bytes.HasSuffix(chunk, promptEnd) {
				return nil
			}
		}
	}
}
