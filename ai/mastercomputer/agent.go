package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/boogie"
	"github.com/theapemachine/amsh/ai/provider"
)

/*
Agent is a type that can communicate with an AI provider and execute operations.
*/
type Agent struct {
	buffer    *Buffer
	processes map[string]Process
	prompt    *Prompt
}

func NewAgent(role string) *Agent {
	return &Agent{
		buffer:    NewBuffer(),
		processes: make(map[string]Process),
		prompt:    NewPrompt(role),
	}
}

/*
Generate uses a simple string as the input and returns a channel of events.
*/
func (agent *Agent) Generate(ctx context.Context, input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		prvdr := provider.NewBalancedProvider()
		accumulator := NewAccumulator()
		accumulator.Stream(
			prvdr.Generate(ctx, provider.GenerationParams{
				Messages: agent.buffer.Truncate(),
			}),
			out,
		)
	}()

	return out
}

/*
Execute the agent for a given operation and state.
*/
func (agent *Agent) Execute(
	ctx context.Context, op boogie.Operation, state boogie.State,
) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		// Prepare the prompt based on operation and context
		agent.buffer.Poke(provider.Message{
			Role:    "user",
			Content: agent.prompt.Build(op, state),
		})

		// Generate using the balanced provider
		prvdr := provider.NewBalancedProvider()
		accumulator := NewAccumulator()
		accumulator.Stream(
			prvdr.Generate(ctx, provider.GenerationParams{
				Messages: agent.buffer.Truncate(),
			}),
			out,
		)

		// Process the result based on operation type
		agent.processes[op.Name].NextState(
			op, accumulator.String(), state,
		)
	}()

	return out
}
