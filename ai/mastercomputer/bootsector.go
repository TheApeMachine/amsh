package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

type BootSector struct {
	ctx   context.Context
	agent *Agent
}

func NewBootSector(ctx context.Context, agent *Agent) *BootSector {
	return &BootSector{ctx: ctx, agent: agent}
}

func (bootsector *BootSector) Startup(input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		accumulator := provider.NewAccumulator()
		accumulator.Stream(bootsector.agent.Generate(input), out)
	}()

	return out
}
