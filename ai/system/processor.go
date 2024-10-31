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

		// Start all cores and forward their events
		for id, core := range p.cores {
			go func(id string, core Core) {
				defer wg.Done()

				// Forward all events in real-time
				for event := range core.team.Execute(input) {
					out <- event
				}
			}(id, core)
		}

		wg.Wait()
	}()

	return out
}
