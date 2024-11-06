package system

import (
	"sync"

	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

type ProcessManager struct {
	key              string
	compositeProcess *process.CompositeProcess
}

func NewProcessManager(key, origin string) *ProcessManager {
	errnie.Info("starting process manager %s %s", key, origin)

	return &ProcessManager{
		key:              key,
		compositeProcess: process.CompositeProcessMap[origin],
	}
}

func (pm *ProcessManager) Execute(accumulator string) <-chan provider.Event {
	errnie.Info("Execute accumulator %s", accumulator)
	out := make(chan provider.Event)

	if pm.compositeProcess == nil || len(pm.compositeProcess.Layers) == 0 {
		errnie.Warn("no composite process found, going for task analysis")
		pm.compositeProcess = process.CompositeProcessMap["task_analysis"]
	}

	go func() {
		defer close(out)

		for _, layer := range pm.compositeProcess.Layers {
			var wg sync.WaitGroup
			wg.Add(len(layer.Processes))

			for event := range NewProcessor(pm.key, layer).Process(accumulator) {
				out <- event

				if event.Type == provider.EventDone {
					wg.Done()
					return
				}
			}

			wg.Wait()
		}

		errnie.Debug("process manager %s completed", pm.key)
	}()

	return out
}
