package system

import (
	"sync"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
)

type ProcessManager struct {
	key              string
	compositeProcess process.CompositeProcess
}

func NewProcessManager(key, origin string) *ProcessManager {
	log.Info("NewProcessManager", "key", key)

	return &ProcessManager{
		key:              key,
		compositeProcess: *process.CompositeProcessMap[key],
	}
}

func (pm *ProcessManager) Execute(accumulator string) <-chan provider.Event {
	log.Info("Execute", "accumulator", accumulator)
	out := make(chan provider.Event)

	if len(pm.compositeProcess.Layers) == 0 {
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
