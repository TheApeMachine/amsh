package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
	"github.com/theapemachine/errnie"
)

/*
System contains all the components of an AI system, loosely modelled on a virtual machine.
It is responsible for kicking off workflows when a workload is received.
*/
type System struct {
	ctx         context.Context
	vm          *VM
	programmers []*Programmer
}

/*
NewSystem creates a new system with a unique key.
*/
func NewSystem(ctx context.Context) *System {
	errnie.Log("system.NewSystem()")

	return &System{ctx: ctx, vm: NewVM(ctx), programmers: make([]*Programmer, 0)}
}

/*
Input kicks off a new workflow with the provided input.
*/
func (system *System) Input(input string) <-chan provider.Event {
	errnie.Log("system.Input(%s)", input)

	out := make(chan provider.Event)

	go func() {
		defer close(out)

		system.programmers = append(system.programmers, NewProgrammer(system.ctx))

		accumulator := provider.NewAccumulator()
		accumulator.Stream(
			system.programmers[0].Generate(input),
			out,
		)

		system.vm.Load(accumulator.String())
	}()

	return out
}

func (system *System) load(input string) {
	errnie.Log("system.load(%s)", input)

	system.vm.Load(utils.StripMarkdown(input, "boogie"))
}
