package marvin

import (
	"context"

	"github.com/theapemachine/amsh/ai/process/fractal"
	"github.com/theapemachine/amsh/ai/process/temporal"
	"github.com/theapemachine/amsh/ai/provider"
)

type System struct{}

func NewSystem() *System {
	return &System{}
}

func (system *System) Process(input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		for _, process := range []Process{
			temporal.NewProcess(),
			fractal.NewProcess(),
		} {
			agent := NewAgent(context.Background(), "reasoner")
			agent.AddProcess(process)
			agent.SetUserPrompt(input)

			accumulator := provider.NewAccumulator()
			accumulator.Stream(agent.Generate(), out)
			input = accumulator.String()
		}
	}()

	return out
}
