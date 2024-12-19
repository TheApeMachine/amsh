package marvin

import (
	"context"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
	"github.com/theapemachine/errnie"
)

type Agent struct {
	Name      string
	Role      string
	Scope     string
	ctx       context.Context
	buffer    *Buffer
	processes map[string]*data.Artifact
	sidekicks map[string][]*Agent
	tool      ai.Tool
	provider  provider.Provider
}

func NewAgent(ctx context.Context, role, scope string, induction *data.Artifact) *Agent {
	return &Agent{
		Name:      utils.NewName(),
		Role:      role,
		Scope:     scope,
		ctx:       ctx,
		buffer:    NewBuffer().Poke(induction),
		processes: make(map[string]*data.Artifact),
		sidekicks: make(map[string][]*Agent),
		provider:  provider.NewBalancedProvider(),
	}
}

func (agent *Agent) AddTool(tool ai.Tool) {
	agent.tool = tool

	message := []string{
		"You have been given access to new tools.",
		"You can use the tools by calling them with the appropriate arguments.",
		"The tools have the following schema:",
		tool.GenerateSchema(),
	}

	agent.buffer.Poke(data.New("assistant", "assistant", "tool", []byte(utils.JoinWith("\n", message...))))
}

func (agent *Agent) AddProcesses(processes ...*data.Artifact) {
	for _, process := range processes {
		agent.processes[process.Peek("role")] = process
	}
}

func (agent *Agent) AddSidekick(key string, sidekick *Agent) {
	agent.sidekicks[key] = append(agent.sidekicks[key], sidekick)

	message := []string{
		"You have been given access to a new sidekick.",
		"You can give the sidekick a task to perform by using the following syntax:",
		"",
		"```json",
		"{",
		"  \"key\": \"" + key + "\",",
		"  \"prompt\": <prompt>",
		"}",
		"```",
		"",
		"Make sure your prompt contains all the necessary information and details to perform the task.",
		"Send the prompt as a continuous string, without any line breaks.",
	}

	agent.buffer.Poke(data.New("assistant", "assistant", "sidekick", []byte(utils.JoinWith("\n", message...))))
}

func (agent *Agent) Generate(prompt *data.Artifact) <-chan *data.Artifact {
	errnie.Info("Generating agent %s %s %s", agent.Name, agent.Role, agent.Scope)

	agent.buffer.Poke(prompt)

	return twoface.NewAccumulator(
		"agent",
		agent.Role,
		agent.Name,
		prompt,
	).Yield(func(accumulator *twoface.Accumulator) {
		defer close(accumulator.Out)

		if agent.Role == "sidekick" {
			agent.handleSidekick(accumulator)
			return
		}

		agent.handleAgent(accumulator)
	}).Generate()
}

func (agent *Agent) handleAgent(accumulator *twoface.Accumulator) {
	for artifact := range agent.provider.Generate(agent.buffer.Peek()) {
		accumulator.Out <- artifact
	}

	blocks := utils.ExtractJSONBlocks(accumulator.Take().Peek("payload"))

	if len(blocks) > 0 {
		for _, block := range blocks {
			// Check if there is a key that matches with the sidekicks.
			for key, sidekicks := range agent.sidekicks {
				if block["key"] == key {
					for _, sidekick := range sidekicks {
						for skArtifact := range sidekick.Generate(
							data.New(agent.Name, "user", "prompt", []byte(block["prompt"].(string))),
						) {
							accumulator.Out <- skArtifact
						}
					}
				}
			}
		}
	}
}

func (agent *Agent) handleSidekick(accumulator *twoface.Accumulator) {
	toolHandler := NewToolHandler(agent)
	toolHandler.Initialize()

	for toolHandler.Accumulator().Take().Peek("payload") != "exit" {
		for artifact := range toolHandler.Accumulator().Generate() {
			accumulator.Out <- artifact
		}
	}
}
