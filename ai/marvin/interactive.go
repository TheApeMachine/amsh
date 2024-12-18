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

// handleInteractiveTool manages the IO stream for interactive tools
func handleInteractiveTool(tool ai.InteractiveTool, agent *Agent, prompt string) <-chan *data.Artifact {
	response := tool.Use(context.Background(), map[string]any{"task": prompt})
	inout := tool.GetIO()
	if inout == nil {
		errnie.Error(fmt.Errorf("tool IO not available"))
		return
	}

	// Generate next command using nested accumulator
	return twoface.NewAccumulator(
		agent.Name,
		"assistant",
		"command",
	).Yield(func(acc *twoface.Accumulator) {
		defer close(acc.Out)

		for artifact := range agent.provider.Generate(agent.buffer.Peek()) {
			acc.Out <- artifact
		}

		command := acc.Take().Peek("payload")
		if command == "" || command == "exit" {
			return
		}

		// Execute command and handle response
		if err := executeCommand(inout, command, acc.Out); err != nil {
			errnie.Error(err)
			return
		}

		// Update message history
		agent.buffer.Poke(acc.Take())
	}).Generate()
}

// Helper function to execute commands and handle responses
func executeCommand(inout io.ReadWriter, command string, out chan<- *data.Artifact) error {
	if _, err := inout.Write([]byte(command + "\n")); err != nil {
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
			n, err := inout.Read(buffer)
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
