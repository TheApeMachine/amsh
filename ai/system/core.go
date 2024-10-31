package system

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
)

type Core struct {
	ctx     context.Context
	cancel  context.CancelFunc
	ID      string
	process process.Process
	team    *ai.Team
	input   chan string
	output  chan process.ProcessResult
}

func NewCore(id string, proc process.Process, key string, outputChan chan process.ProcessResult) Core {
	log.Info("NewCore", "id", id, "key", key)

	ctx, cancel := context.WithCancel(context.Background())

	return Core{
		ctx:     ctx,
		cancel:  cancel,
		ID:      id,
		process: proc,
		team:    ai.NewTeam(ctx, key, proc.SystemPrompt(key)),
		input:   make(chan string, 1),
		output:  outputChan,
	}
}

func (core *Core) Run() chan process.ProcessResult {
	log.Info("Starting core", "id", core.ID)
	outputChan := make(chan process.ProcessResult, 1)

	go func() {
		defer close(outputChan)

		for input := range core.input {
			log.Info("Core received input", "id", core.ID, "input", input)

			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			var result process.ProcessResult
			result.CoreID = core.ID

			// Create done channel for agent execution
			done := make(chan struct{})
			var output string

			// Execute agent in separate goroutine
			go func() {
				defer close(done)
				for event := range core.team.Execute(input) {
					if event.Type == provider.EventToken {
						output += event.Content
					}
				}
			}()

			// Wait for either completion or timeout
			select {
			case <-done:
				if len(output) > 0 {
					result.Data = json.RawMessage(output)
				} else {
					result.Error = fmt.Errorf("empty response from agent")
				}
			case <-ctx.Done():
				result.Error = fmt.Errorf("timeout waiting for agent response")
			}

			cancel()
			select {
			case outputChan <- result:
			default:
				log.Error("Failed to send result", "core", core.ID)
			}
		}
	}()

	return outputChan
}
