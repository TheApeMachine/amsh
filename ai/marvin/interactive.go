package marvin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

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
	if _, err := toolHandler.inout.Write([]byte(command + "\n")); err != nil {
		return err
	}

	buffer := make([]byte, 4096)
	var response []byte

	// Read with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan bool)
	go func() {
		for {
			n, err := toolHandler.inout.Read(buffer)
			if err != nil && err != io.EOF {
				errnie.Error(err)
				done <- true
				return
			}
			if n > 0 {
				response = append(response, buffer[:n]...)
				out <- data.New("tool", "output", "stream", buffer[:n])
			}
			if bytes.HasSuffix(response, []byte("# ")) {
				done <- true
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
