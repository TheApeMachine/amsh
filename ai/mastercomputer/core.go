package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/boogie"
	"github.com/theapemachine/amsh/ai/provider"
)

type Core struct {
	agent *Agent
}

func NewCore(role string) *Core {
	return &Core{
		agent: NewAgent(role),
	}
}

func (core *Core) Execute(
	ctx context.Context, op boogie.Operation, state boogie.State,
) <-chan provider.Event {
	// Delegate execution to agent
	return core.agent.Execute(ctx, op, state)
}
