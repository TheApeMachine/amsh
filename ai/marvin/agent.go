package marvin

import (
	"context"

	"github.com/google/uuid"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

/*
Agent is a type that can communicate with an AI provider and execute operations.
*/
type Agent struct {
	ID        string
	ctx       context.Context
	buffer    *Buffer
	processes map[string]Process
	sidekicks map[string][]*Agent
	prompt    *Prompt
	role      string
}

func NewAgent(ctx context.Context, role string) *Agent {
	return &Agent{
		ID:        uuid.New().String(),
		ctx:       ctx,
		buffer:    NewBuffer(),
		processes: make(map[string]Process),
		sidekicks: make(map[string][]*Agent),
		prompt:    NewPrompt(role),
		role:      role,
	}
}

// GetBuffer returns the agent's buffer
func (agent *Agent) GetBuffer() *Buffer {
	return agent.buffer
}

// GetPrompt returns the agent's prompt
func (agent *Agent) GetPrompt() *Prompt {
	return agent.prompt
}

// GetRole returns the agent's role
func (agent *Agent) GetRole() string {
	return agent.role
}

// GetContext returns the agent's context
func (agent *Agent) GetContext() context.Context {
	return agent.ctx
}

func (agent *Agent) AddProcess(process Process) {
	agent.prompt.AddProcess(process)
}

func (agent *Agent) AddSidekick(key string, sidekick *Agent) {
	agent.sidekicks[key] = append(agent.sidekicks[key], sidekick)
}

func (agent *Agent) SetUserPrompt(userPrompt string) {
	agent.prompt.SetUserPrompt(userPrompt)
}

/*
Generate uses a simple string as the input and returns a channel of events.
*/
func (agent *Agent) Generate() <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		// Basic prompt setup
		agent.buffer.Clear().
			Poke(agent.prompt.System()).
			Poke(agent.prompt.User()).
			Poke(agent.prompt.Context())

		errnie.Log("%s", agent.buffer.Truncate())

		// Generate response
		prvdr := provider.NewBalancedProvider()

		accumulator := provider.NewAccumulator()
		accumulator.Stream(prvdr.Generate(agent.ctx, provider.GenerationParams{
			Messages: agent.buffer.Truncate(),
		}), out)

		errnie.Log("%s", accumulator.String())
	}()

	return out
}
