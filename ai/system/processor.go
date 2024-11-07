package system

import (
	"context"

	"github.com/theapemachine/amsh/ai/process/layering"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
)

/*
Processor is a struct that manages the cores and channels for a distributed AI system.
*/
type Processor struct {
	ctx context.Context
	key string
}

/*
NewProcessor creates a new processor with the given key.
*/
func NewProcessor(ctx context.Context, key string) *Processor {
	errnie.Info("NewProcessor %s", key)

	return &Processor{
		ctx: ctx,
		key: key,
	}
}

/*
Process processes the input string and returns a channel of results.
*/
func (processor *Processor) Process(layer layering.Layer) <-chan provider.Event {
	errnie.Info("Processor.Process %s", layer)
	out := make(chan provider.Event)

	go func() {
		defer close(out)

		for _, workload := range layer.Workloads {
			for event := range NewCore(processor.ctx, processor.key).Run(workload) {
				out <- event
			}
		}

		errnie.Debug("processor %s completed", processor.key)
	}()

	return out
}
