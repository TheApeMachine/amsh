package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/boogie"
	"github.com/theapemachine/amsh/ai/provider"
)

type BootSector struct {
	agent *Agent
}

func NewBootSector(agent *Agent) *BootSector {
	return &BootSector{agent: agent}
}

func (bootsector *BootSector) Startup(ctx context.Context, input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		bootsector.agent.Execute(ctx, boogie.Operation{
			Name: "bootsector",
		}, boogie.State{})
	}()

	return out
}
