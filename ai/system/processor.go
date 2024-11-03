package system

import (
	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/process"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

/*
Processor is a struct that manages the cores and channels for a distributed AI system.
*/
type Processor struct {
	key   string
	layer *process.Layer
}

/*
NewProcessor creates a new processor with the given key.
*/
func NewProcessor(key string, layer *process.Layer) *Processor {
	log.Info("NewProcessor", "key", key)

	return &Processor{
		key:   key,
		layer: layer,
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
			for event := range NewCore(processor.key, process).Run(input) {
				out <- event
			}
		}

		errnie.Debug("processor %s completed", processor.key)
	}()

	return out
}
