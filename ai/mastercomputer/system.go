package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

type System struct {
	key string
}

func NewSystem(key string) *System {
	return &System{key: key}
}

func (system *System) Input(input string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		provider.NewAccumulator().Stream(NewBootSector(
			NewAgent(system.key, "bootsector"),
		).Startup(ctx, input), out)
	}()

	return out
}
