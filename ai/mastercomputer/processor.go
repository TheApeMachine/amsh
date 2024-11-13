package mastercomputer

import (
	"context"

	"github.com/theapemachine/amsh/ai/provider"
)

type Processor struct {
	cores map[string]*Core
}

func NewProcessor() *Processor {
	return &Processor{
		cores: make(map[string]*Core),
	}
}

func (processor *Processor) Execute(
	ctx context.Context, prompt string,
) <-chan provider.Event {
	out := make(chan provider.Event)
	return out
}
