package system

import (
	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
)

type ProcessManager struct {
	key       string
	toolset   *ai.Toolset
	processor *Processor
}

func NewProcessManager(key, origin string) *ProcessManager {
	log.Info("NewProcessManager", "key", key)

	pm := &ProcessManager{
		key:       key,
		toolset:   ai.NewToolset(),
		processor: NewProcessor(key),
	}

	return pm
}

func (pm *ProcessManager) Execute(incoming string) <-chan provider.Event {
	log.Info("Execute", "incoming", incoming)

	out := make(chan provider.Event)

	go func() {
		defer close(out)

		// Process through all cores
		log.Info("Starting processing", "incoming", incoming)
		results := pm.processor.Process(incoming)

		// Collect and combine results
		var finalResult process.ThinkingResult
		for result := range results {
			log.Info("Received result", "core", result.CoreID)
			if result.Error != nil {
				log.Error("Core error", "core", result.CoreID, "error", result.Error)
				out <- provider.Event{
					Type:    provider.EventError,
					Content: result.Error.Error(),
				}
				return
			}

			if err := finalResult.Integrate(result); err != nil {
				log.Error("Integration error", "error", err)
				out <- provider.Event{
					Type:    provider.EventError,
					Content: err.Error(),
				}
				return
			}
		}

		// Send integrated result
		output, err := json.Marshal(finalResult)
		if err != nil {
			log.Error("Marshal error", "error", err)
			out <- provider.Event{
				Type:    provider.EventError,
				Content: err.Error(),
			}
			return
		}

		out <- provider.Event{
			Type:    provider.EventToken,
			Content: string(output),
		}
	}()

	return out
}
