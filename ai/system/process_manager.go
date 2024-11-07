package system

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process/layering"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type ProcessManager struct {
	ctx     context.Context
	cancel  context.CancelFunc
	key     string
	manager *ai.Agent
}

func NewProcessManager(key, origin string) *ProcessManager {
	errnie.Info("starting process manager %s %s", key, origin)
	ctx, cancel := context.WithCancel(context.Background())

	return &ProcessManager{
		ctx:    ctx,
		cancel: cancel,
		key:    key,
		manager: ai.NewAgent(
			ctx, key, "layering", "manager", layering.NewProcess().SystemPrompt(key), nil,
		),
	}
}

func (pm *ProcessManager) Execute(request string) <-chan provider.Event {
	errnie.Info("Execute request %s", request)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		var layerAccumulator string

		for event := range pm.manager.Execute(request) {
			layerAccumulator += event.Content
			out <- event
		}

		accumulators := make(map[int]string)

		if process := pm.validate(layerAccumulator); process != nil {
			for idx, layer := range process.Layers {
				errnie.Info("executing layer %s", layer.Workloads)

				var wg sync.WaitGroup
				wg.Add(len(layer.Workloads))

				ctx, cancel := context.WithCancel(pm.ctx)
				defer cancel()

				for event := range NewProcessor(ctx, pm.key).Process(layer) {
					accumulators[idx] += event.Content
					out <- event
				}

				wg.Wait()
			}
		}

		errnie.Debug("process manager %s completed", pm.key)
	}()

	return out
}

func (pm *ProcessManager) validate(accumulator string) *layering.Process {
	process := layering.NewProcess()
	errnie.MustVoid(json.Unmarshal([]byte(utils.StripMarkdown(accumulator)), process))
	return process
}
