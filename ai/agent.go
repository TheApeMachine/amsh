package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
)

// AgentState represents the current state of an agent
type AgentState string

const (
	StateIdle    AgentState = "idle"
	StateWorking AgentState = "working"
	StateWaiting AgentState = "waiting"
	StateDone    AgentState = "done"
)

// Agent represents an AI agent that can perform tasks and communicate with other agents
type Agent struct {
	Name     string  `json:"name"`
	Role     string  `json:"role"`
	Buffer   *Buffer `json:"buffer"`
	provider provider.Provider
	state    AgentState
}

// NewAgent creates a new agent with integrated reasoning and learning
func NewAgent(key, role, systemPrompt string) *Agent {
	log.Info("NewAgent", "key", key, "role", role)
	return &Agent{
		Name: fmt.Sprintf("%s-%s", key, role),
		Role: role,
		Buffer: NewBuffer().AddMessage("system", strings.Join([]string{
			systemPrompt,
			viper.GetString(fmt.Sprintf("ai.setups.%s.agent.prompt", key)),
		}, "\n")),
		provider: provider.NewBalancedProvider(),
		state:    StateIdle,
	}
}

func (agent *Agent) Execute(prompt string) <-chan provider.Event {
	log.Info("executing agent", "agent", agent.Name, "prompt", prompt)
	out := make(chan provider.Event)

	agent.Buffer.AddMessage("user", strings.Join([]string{
		"<prompt>",
		prompt,
		"</prompt>",
	}, "\n"))

	go func() {
		defer close(out)

		var accumulator string

		for event := range agent.provider.Generate(context.Background(), agent.Buffer.GetMessages()) {
			accumulator += event.Content
			out <- event
		}

		out <- provider.Event{
			Type:    provider.EventDone,
			Content: "\n",
		}

		accumulator += "\n"
		agent.Buffer.AddMessage("assistant", accumulator)
	}()

	return out
}
