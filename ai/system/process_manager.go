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

			// Create a processor for this layer
			processor := NewProcessor(pm.key, layer...)

			// Start processing and collect results
			resultChan := processor.Process(accumulator)

			// Track active cores
			wg.Add(1)
			go func() {
				defer wg.Done()
				for result := range resultChan {
					if result.Type == provider.EventToken {
						accumulator += result.Content
					}
					out <- result
				}
			}()

			// Wait for all processing to complete before moving to next layer
			wg.Wait()
			log.Info("Layer processing complete", "layer", layer)
		}
	}()

	return out
}
