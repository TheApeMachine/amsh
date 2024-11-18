package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/qpool"
)

type Core struct {
	ctx  context.Context
	pool *qpool.Q
}

func NewCore(ctx context.Context, pool *qpool.Q) *Core {
	return &Core{
		ctx:  ctx,
		pool: pool,
	}
}

func (core *Core) Generate(in string) <-chan provider.Event {
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		value := <-core.pool.Schedule(in, func() (any, error) {
			accumulator := provider.NewAccumulator()
			return accumulator.Collect(
				NewAgent(core.ctx, "worker").Generate(in),
			), nil
		})

		out <- provider.Event{
			Content: value.Value.(string),
		}
	}()

	return out
}
