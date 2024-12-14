package mastercomputer

import (
	"context"

	"github.com/google/uuid"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/errnie"
)

/*
Agent is a type that can communicate with an AI provider and execute operations.
*/
type Agent struct {
	ID        string
	ctx       context.Context
	buffer    *Buffer
	processes map[string]Process
	prompt    *Prompt
}

func NewAgent(ctx context.Context, role string) *Agent {
	return &Agent{
		ID:        uuid.New().String(),
		ctx:       ctx,
		buffer:    NewBuffer(),
		processes: make(map[string]Process),
		prompt:    NewPrompt(role),
	}
}

/*
Generate uses a simple string as the input and returns a channel of events.
*/
func (agent *Agent) Generate(input string) <-chan provider.Event {
	errnie.Log("%s", input)
	out := make(chan provider.Event)

	agent.buffer.Clear().Poke(agent.prompt.System()).Poke(provider.Message{
		Role:    "user",
		Content: input,
	})

	errnie.Log("%s", agent.buffer.Truncate())

	go func() {
		defer close(out)

		prvdr := provider.NewBalancedProvider()
		accumulator := provider.NewAccumulator()
		accumulator.Stream(prvdr.Generate(agent.ctx, provider.GenerationParams{
			Messages: agent.buffer.Truncate(),
		}), out)

		errnie.Log("%s", accumulator.String())
	}()

	return out
}
