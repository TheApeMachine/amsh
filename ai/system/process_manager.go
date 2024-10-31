package system

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/provider"
)

type ProcessManager struct {
	key string
}

func NewProcessManager(key, origin string) *ProcessManager {
	log.Info("NewProcessManager", "key", key)

	pm := &ProcessManager{
		key: key,
	}

	return pm
}

func (pm *ProcessManager) Execute(accumulator string) <-chan provider.Event {
	log.Info("Execute", "accumulator", accumulator)
	out := make(chan provider.Event)

	// Handle terminal resizing
	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)

	go func() {
		defer close(out)

		for _, layer := range [][]string{
			{"surface", "pattern", "quantum", "time"},
			{"narrative", "analogy", "practical", "context"},
			{"moonshot", "sensible", "catalyst", "guardian"},
			{"performance", "memory", "oversight", "integration"},
			{"programmer", "data_scientist", "qa_engineer", "security_specialist"},
		} {
			var wg sync.WaitGroup
			wg.Add(len(layer))

			log.Info("Starting processing", "layer", layer)

			for result := range NewProcessor(
				pm.key, layer...,
			).Process(accumulator) {
				if result.Type == provider.EventDone {
					wg.Done()
				}

				if result.Type == provider.EventToken {
					accumulator += result.Content
				}

				out <- result
			}

			wg.Wait()
		}
	}()

	return out
}
