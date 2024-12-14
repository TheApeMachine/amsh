package system

import (
	"context"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process/layering"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/errnie"
)

type Core struct {
	ctx context.Context
	key string
}

func NewCore(ctx context.Context, key string) *Core {
	errnie.Info("new core created %s", key)
	return &Core{
		ctx: ctx,
		key: key,
	}
}

func (core *Core) Run(workload layering.Workload) <-chan provider.Event {
	errnie.Info("Starting core %s", core.key)
	out := make(chan provider.Event, 1)

	go func() {
		defer close(out)

		for event := range ai.NewTeam(
			core.ctx, core.key,
		).Execute(workload) {
			out <- event
		}

		errnie.Debug("core %s completed", core.key)
	}()

	return out
}
