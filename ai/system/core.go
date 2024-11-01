package system

import (
	"context"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
)

type Core struct {
	ctx     context.Context
	cancel  context.CancelFunc
	key     string
	process process.Process
	team    *ai.Team
	wg      *sync.WaitGroup
}

func NewCore(key string, proc process.Process, wg *sync.WaitGroup) *Core {
	log.Info("NewCore", "key", key)

	ctx, cancel := context.WithCancel(context.Background())

	return &Core{
		ctx:     ctx,
		cancel:  cancel,
		key:     key,
		process: proc,
		team:    ai.NewTeam(ctx, key, "", proc),
		wg:      wg,
	}
}

func (core *Core) Run() <-chan provider.Event {
	log.Info("Starting core", "key", core.key)
	out := make(chan provider.Event, 1)

	go func() {
		defer close(out)
	}()

	return out
}
