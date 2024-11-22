package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/boogie"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/qpool"
)

type Processor struct {
	ctx         context.Context
	pool        *qpool.Q
	instruction boogie.Instruction
	cores       []*Core
}

func NewProcessor(ctx context.Context, instruction boogie.Instruction) *Processor {
	errnie.Log("processor.NewProcessor(%v)", instruction)

	return &Processor{
		ctx:         ctx,
		pool:        qpool.NewQ(ctx, 1, 4, &qpool.Config{}),
		instruction: instruction,
		cores:       make([]*Core, 0),
	}
}

func (processor *Processor) Generate(in string) chan provider.Event {
	errnie.Log("processor.Generate(%s)", in)

	out := make(chan provider.Event)

	go func() {
		defer close(out)

		processor.cores = append(processor.cores, NewCore(processor.ctx, processor.pool))

		for _, core := range processor.cores {
			accumulator := provider.NewAccumulator()
			accumulator.Stream(
				core.Generate(in),
				out,
			)
		}
	}()

	return out
}
