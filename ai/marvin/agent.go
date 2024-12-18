package marvin

import (
	"context"
	"io"

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
	tools     []ai.Tool
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
		tools:     make([]ai.Tool, 0),
		provider:  provider.NewBalancedProvider(),
	}
}

func (agent *Agent) AddTools(tools ...ai.Tool) {
	agent.tools = append(agent.tools, tools...)

	message := []string{
		"You have been given access to new tools.",
		"You can use the tools by calling them with the appropriate arguments.",
		"The tools have the following schema:",
	}

	for _, tool := range tools {
		message = append(message, tool.GenerateSchema())
	}

	agent.buffer.Poke(data.New("assistant", "assistant", "tools", []byte(utils.JoinWith("\n", message...))))
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
	}

	agent.buffer.Poke(data.New("assistant", "assistant", "sidekick", []byte(utils.JoinWith("\n", message...))))
}

func (agent *Agent) Generate(prompt *data.Artifact) <-chan *data.Artifact {
	errnie.Info("Generating agent", "agent", agent.Name, "role", agent.Role, "scope", agent.Scope)

	return twoface.NewAccumulator(
		"agent",
		agent.Role,
		agent.Name,
		prompt,
	).Yield(func(accumulator *twoface.Accumulator) {
		messages := agent.buffer.Poke(prompt).Peek()
		defer close(accumulator.Out)

		// For sidekicks with tools, skip initial generation and go straight to tool interaction
		if agent.Role == "sidekick" && len(agent.tools) > 0 {
			for _, tool := range agent.tools {
				if interactiveTool, ok := tool.(ai.InteractiveTool); ok {
					handleInteractiveTool(interactiveTool, agent, string(prompt.Peek("payload")), accumulator.Out)
				}
			}
			return
		}

		// Normal agent flow
		buffer := agent.streamOutput(messages, accumulator.Out)

		// Handle tools for non-sidekick agents
		if len(agent.tools) > 0 {
			for _, tool := range agent.tools {
				if interactiveTool, ok := tool.(ai.InteractiveTool); ok {
					handleInteractiveTool(interactiveTool, agent, string(buffer), accumulator.Out)

					if inout := interactiveTool.GetIO(); inout != nil {
						buffer := make([]byte, 1024)
						for {
							n, err := inout.Read(buffer)
							if err != nil {
								if err != io.EOF {
									errnie.Error(err)
								}
								break
							}

							toolResponse := data.New(
								agent.Name,
								"tool",
								"response",
								buffer[:n],
							)

							for artifact := range agent.provider.Generate(append(messages, toolResponse)) {
								accumulator.Out <- artifact
							}
						}
					}
				}
			}
		}

		agent.processSidekickCalls(buffer, accumulator.Out)
		agent.buffer.Poke(accumulator.Take())
	}).Generate()
}

// streamOutput handles the main generation and returns the accumulated buffer
func (agent *Agent) streamOutput(messages []*data.Artifact, out chan<- *data.Artifact) []byte {
	var buffer []byte
	for artifact := range agent.provider.Generate(messages) {
		buffer = append(buffer, artifact.Peek("payload")...)
		out <- artifact
	}
	return buffer
}

// processSidekickCalls handles executing any sidekick tasks found in the output
func (agent *Agent) processSidekickCalls(buffer []byte, out chan<- *data.Artifact) {
	blocks := utils.ExtractJSONBlocks(string(buffer))
	for _, block := range blocks {
		agent.executeSidekickCall(block, out)
	}
}

// executeSidekickCall processes a single sidekick call if valid
func (agent *Agent) executeSidekickCall(block map[string]interface{}, out chan<- *data.Artifact) {
	key, ok := block["key"].(string)
	if !ok {
		return
	}

	sidekicks, exists := agent.sidekicks[key]
	if !exists {
		return
	}

	prompt, ok := block["prompt"].(string)
	if !ok {
		return
	}

	for _, sidekick := range sidekicks {
		agent.forwardSidekickResponse(sidekick, prompt, out)
	}
}

// forwardSidekickResponse generates and forwards a sidekick's response
func (agent *Agent) forwardSidekickResponse(sidekick *Agent, prompt string, out chan<- *data.Artifact) {
	task := data.New(
		agent.Name,
		"user",
		"task",
		[]byte(prompt),
	)

	// Create accumulator for sidekick response
	sidekickAccumulator := twoface.NewAccumulator(
		"sidekick",
		sidekick.Role,
		sidekick.Name,
		task,
	).Yield(func(accumulator *twoface.Accumulator) {
		defer close(accumulator.Out)

		// Check if sidekick has an interactive tool
		if tool, ok := sidekick.tools[0].(ai.InteractiveTool); ok && tool.IsInteractive() {
			handleInteractiveTool(tool, sidekick, prompt, accumulator.Out)
			return
		}

		// Handle normal tool response
		for artifact := range sidekick.Generate(task) {
			accumulator.Out <- artifact
		}
	})

	// Forward accumulated responses to main output
	for artifact := range sidekickAccumulator.Generate() {
		out <- artifact
	}

	// Update agent's buffer with sidekick response
	agent.buffer.Poke(sidekickAccumulator.Take())
}
