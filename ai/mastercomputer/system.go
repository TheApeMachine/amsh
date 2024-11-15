package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

/*
System contains all the components of an AI system, loosely modelled on a virtual machine.
It is responsible for kicking off workflows when a workload is received.
*/
type System struct {
	ctx context.Context
}

/*
NewSystem creates a new system with a unique key.
*/
func NewSystem(ctx context.Context) *System {
	return &System{ctx: ctx}
}

/*
Input kicks off a new workflow with the provided input.
*/
func (system *System) Input(input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		provider.NewAccumulator().Stream(NewBootSector(
			system.ctx,
			NewAgent(system.ctx, "bootsector"),
		).Startup(input), out)
	}()

	return out
}
