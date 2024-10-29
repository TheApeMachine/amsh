package system

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process"
)

type Core struct {
	id      string
	process process.Process
	agent   *ai.Agent
	input   chan string
	output  chan process.ProcessResult
}

func NewCore(id string, proc process.Process, key string, outputChan chan process.ProcessResult) Core {
	log.Info("NewCore", "id", id, "key", key)

	return Core{
		id:      id,
		process: proc,
		agent: ai.NewAgent(
			fmt.Sprintf("%s-%s", key, id),
			id,
			proc.SystemPrompt(key),
		),
		input:  make(chan string, 1),
		output: outputChan,
	}
}

func (c *Core) Run() {
	log.Info("Starting core", "id", c.id)

	for input := range c.input {
		log.Info("Core received input", "id", c.id, "input", input)
		var result process.ProcessResult
		result.CoreID = c.id

		// Execute the agent with the input
		output := ""
		for event := range c.agent.Execute(input) {
			output += event.Content
		}

		// Parse the output
		result.Data = json.RawMessage(output)

		// Send to the processor's channel, not the core's output
		c.output <- result
		log.Info("Core sent result", "id", c.id)
	}
}
