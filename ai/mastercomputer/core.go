package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

type Core struct {
	ctx   context.Context
	agent *Agent
}

func NewCore(ctx context.Context, role string) *Core {
	return &Core{
		ctx:   ctx,
		agent: NewAgent(ctx, role),
	}
}

func (core *Core) Execute(prompt string) <-chan provider.Event {
	return core.agent.Generate(prompt)
}
