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
	key   string
	layer *process.Layer
	wg    *sync.WaitGroup
}

/*
NewProcessor creates a new processor with the given key.
*/
func NewProcessor(key string, layer *process.Layer, wg *sync.WaitGroup) *Processor {
	log.Info("NewProcessor", "key", key)

	return &Processor{
		key:   key,
		layer: layer,
		wg:    wg,
	}
}

/*
Process processes the input string and returns a channel of results.
*/
func (processor *Processor) Process(input string) <-chan provider.Event {
	log.Info("Processor.Process", "input", input)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		for _, process := range processor.layer.Processes {
			out <- <-NewCore(processor.key, process, processor.wg).Run(input)
		}
	}()

	return out
}
