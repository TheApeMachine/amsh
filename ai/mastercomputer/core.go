package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

type Core struct {
	key   string
	agent *Agent
}

func NewCore(key, role string) *Core {
	return &Core{
		key:   key,
		agent: NewAgent(key, role),
	}
}

func (core *Core) Execute(
	ctx context.Context, prompt string,
) <-chan provider.Event {
	return core.agent.Generate(ctx, prompt)
}
