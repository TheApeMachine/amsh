package system

import (
	"fmt"

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

	go func() {
		defer close(out)

		for _, cores := range [][]string{
			{"surface", "pattern", "quantum", "time"},
			{"narrative", "analogy", "practical", "context"},
			{"moonshot", "sensible", "catalyst", "guardian"},
			{"programmer", "data_scientist", "qa_engineer", "security_specialist"},
		} {
			log.Info("Starting processing", "cores", cores)

			for result := range NewProcessor(
				pm.key, cores...,
			).Process(accumulator) {
				if result.Type == provider.EventToken {
					accumulator += result.Content
				}

				out <- result
			}

			pm.memory.Execute(accumulator)
		}
	}()

	return out
}
