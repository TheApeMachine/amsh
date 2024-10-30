package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"golang.org/x/exp/rand"
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
	params   provider.GenerationParams
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

		for event := range agent.provider.Generate(
			context.Background(), agent.params, agent.Buffer.GetMessages(),
		) {
			if event.Type == provider.EventToken {
				accumulator += event.Content
				out <- event
			}
		}

		agent.Buffer.AddMessage("assistant", accumulator)
		agent.Tweak()
	}()

	return out
}

func (agent *Agent) Tweak() provider.GenerationParams {
	agent.params.Interestingness = agent.MeasureInterestingness()
	if len(agent.params.InterestingnessHistory) > 5 {
		// If results getting boring, increase temperature
		if average(agent.params.InterestingnessHistory) < 0.5 {
			agent.params.Temperature *= 1.1
			agent.params.TopK += 10
		}

		// If results too wild, decrease temperature
		if average(agent.params.InterestingnessHistory) > 0.8 {
			agent.params.Temperature *= 0.9
			agent.params.TopK -= 5
		}

		// Keep a moving window
		agent.params.InterestingnessHistory = agent.params.InterestingnessHistory[1:]
	}

	return agent.params
}

func (agent *Agent) MeasureInterestingness() float64 {
	interestingness := measureInterestingness()
	agent.params.InterestingnessHistory = append(
		agent.params.InterestingnessHistory, interestingness,
	)

	return interestingness
}

func measureInterestingness() float64 {
	return rand.Float64()
}

func average(values []float64) float64 {
	return sum(values) / float64(len(values))
}

func sum(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total
}
