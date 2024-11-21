package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

type Programmer struct {
	agent *Agent
}

func NewProgrammer(ctx context.Context) *Programmer {
	return &Programmer{
		agent: NewAgent(ctx, "programmer"),
	}
}

func (programmer *Programmer) Generate(input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)
		provider.NewAccumulator().Stream(programmer.agent.Generate(input), out)
	}()

	return out
}
