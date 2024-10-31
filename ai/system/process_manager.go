package system

import (
	"encoding/json"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/process"
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

	go func() {
		defer close(out)

		// Create processor for task analysis
		processor := NewProcessor(pm.key, "task_analyzer")

		// Get analysis results
		resultChan := processor.Process(accumulator)
		var analysis process.TaskAnalysis
		var analysisStr string

		for result := range resultChan {
			if result.Type == provider.EventToken {
				analysisStr += result.Content
			}
		}

		if err := json.Unmarshal([]byte(analysisStr), &analysis); err != nil {
			log.Error("Failed to parse task analysis", "error", err)
			return
		}

		// Process each required layer group in priority order
		for _, group := range analysis.RequiredLayers {
			var wg sync.WaitGroup
			processor := NewProcessor(pm.key, group.Layers...)
			resultChan := processor.Process(accumulator)

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

			wg.Wait()
			log.Info("Layer group processing complete",
				"group", group.Name,
				"layers", group.Layers)
		}
	}()

	return out
}
