package system

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/provider"
)

type ProcessManager struct {
	key     string
	toolset *ai.Toolset
	memory  *ai.Agent
}

func NewProcessManager(key, origin string) *ProcessManager {
	log.Info("NewProcessManager", "key", key)

	pm := &ProcessManager{
		key:     key,
		toolset: ai.NewToolset(),
		memory: ai.NewAgent(key, origin, viper.GetViper().GetString(
			fmt.Sprintf("ai.setups.%s.memory.prompt", key),
		)),
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

			pm.memory.Execute(accumulator)

			wg.Wait()
		}
	}()

	return out
}
