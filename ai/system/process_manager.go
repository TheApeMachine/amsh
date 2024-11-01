package system

import (
	"sync"

	"github.com/charmbracelet/log"
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
	log.Info("Execute", "accumulator", accumulator)
	out := make(chan provider.Event)

	if pm.compositeProcess == nil || len(pm.compositeProcess.Layers) == 0 {
		log.Error("no composite process found, going for task analysis")
	}

	go func() {
		defer close(out)

		for _, layer := range pm.compositeProcess.Layers {
			var wg sync.WaitGroup
			wg.Add(len(layer.Processes))

			for event := range NewProcessor(pm.key, layer, &wg).Process(accumulator) {
				out <- event
			}

			wg.Wait()
		}
	}()

	return out
}
