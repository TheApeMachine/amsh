package system

import (
	"sync"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
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
func NewProcessor(key string, coreTypes ...string) *Processor {
	log.Info("NewProcessor", "key", key)

	p := &Processor{
		cores:    make(map[string]Core),
		channels: make(map[string]chan process.ProcessResult),
		done:     make(chan struct{}),
	}

	// Initialize channels before creating cores
	for _, id := range coreTypes {
		p.channels[id] = make(chan process.ProcessResult, 1)
	}

	// Create cores with their respective channels
	for _, id := range coreTypes {
		p.cores[id] = NewCore(id, process.ProcessMap[id], key, p.channels[id])
	}

	return p
}

/*
Process processes the input string and returns a channel of results.
*/
func (p *Processor) Process(input string) <-chan provider.Event {
	log.Info("Processor.Process", "input", input)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		var wg sync.WaitGroup
		wg.Add(len(p.cores))

		// Start all cores
		for id, core := range p.cores {
			go func(id string, core Core) {
				defer wg.Done()

				// Start the core's Run method
				resultChan := core.Run()

				// Send input to the core
				core.input <- input

				// Process results
				for result := range resultChan {
					if result.Error != nil {
						out <- provider.Event{
							Type:    provider.EventError,
							Content: result.Error.Error(),
							AgentID: id,
						}
						continue
					}

					// Convert result data to events
					if len(result.Data) > 0 {
						out <- provider.Event{
							Type:    provider.EventToken,
							Content: string(result.Data),
							AgentID: id,
						}
					}
				}
			}(id, core)
		}

		wg.Wait()
	}()

	return out
}
