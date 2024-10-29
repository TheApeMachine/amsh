package system

import (
	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/process"
)

/*
Processor is a struct that manages the cores and channels for a distributed AI system.
*/
type Processor struct {
	cores    map[string]Core
	channels map[string]chan process.ProcessResult
	done     chan struct{}
}

/*
NewProcessor creates a new processor with the given key.
*/
func NewProcessor(key string) *Processor {
	log.Info("NewProcessor", "key", key)

	p := &Processor{
		channels: make(map[string]chan process.ProcessResult),
		done:     make(chan struct{}),
	}

	// First create the channels
	p.channels = map[string]chan process.ProcessResult{
		"surface": make(chan process.ProcessResult, 1),
		"pattern": make(chan process.ProcessResult, 1),
		"quantum": make(chan process.ProcessResult, 1),
		"time":    make(chan process.ProcessResult, 1),
	}

	// Then create cores with buffered input channels
	p.cores = map[string]Core{
		"surface": NewCore("surface", process.NewSurfaceAnalysis(), key, p.channels["surface"]),
		"pattern": NewCore("pattern", process.NewPatternAnalysis(), key, p.channels["pattern"]),
		"quantum": NewCore("quantum", process.NewQuantumAnalysis(), key, p.channels["quantum"]),
		"time":    NewCore("time", process.NewTimeAnalysis(), key, p.channels["time"]),
	}

	// Start all cores in goroutines
	for id, core := range p.cores {
		go func(id string, c Core) {
			log.Info("Starting core", "id", id)
			c.Run()
		}(id, core)
	}

	return p
}

/*
Process processes the input string and returns a channel of results.
*/
func (p *Processor) Process(input string) <-chan process.ProcessResult {
	log.Info("Process", "input", input)
	results := make(chan process.ProcessResult)

	go func() {
		defer close(results)

		// Send input to all cores sequentially
		for id, core := range p.cores {
			log.Info("Sending to core", "id", id)
			select {
			case core.input <- input:
				// Wait for result from this core
				if result := <-p.channels[id]; result.Error != nil {
					results <- result
					return
				} else {
					results <- result
				}
			default:
				log.Error("Failed to send to core", "id", id)
			}
		}
	}()

	return results
}
