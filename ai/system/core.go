package system

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

type Core struct {
	ctx     context.Context
	cancel  context.CancelFunc
	key     string
	process process.Process
}

func NewCore(key string, proc process.Process) *Core {
	errnie.Info("new core created %s", key)

	ctx, cancel := context.WithCancel(context.Background())

	return &Core{
		ctx:     ctx,
		cancel:  cancel,
		key:     key,
		process: proc,
	}
}

func (core *Core) Run(input string) <-chan provider.Event {
	log.Info("Starting core", "key", core.key)
	out := make(chan provider.Event, 1)

	go func() {
		defer close(out)

		for event := range ai.NewTeam(
			core.ctx, core.key, core.process,
		).Execute(input) {
			out <- event
		}

		errnie.Debug("core %s completed", core.key)
	}()

	return out
}
