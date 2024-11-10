package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

type System struct {
}

func NewSystem() *System {
	return &System{}
}

func (system *System) Input(input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		NewAccumulator().Stream(
			NewBootSector(
				NewAgent("bootsector"),
			).Startup(ctx, input),
			out,
		)
	}()

	return out
}
