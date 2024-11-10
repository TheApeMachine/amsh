package mastercomputer

import (
	"strings"
	"sync"

	"github.com/theapemachine/amsh/ai/provider"
)

type Accumulator struct {
	buffer *sync.Map
}

func NewAccumulator() *Accumulator {
	return &Accumulator{buffer: &sync.Map{}}
}

func (accumulator *Accumulator) Stream(
	in <-chan provider.Event,
	out chan<- provider.Event,
) {
	for event := range in {
		if _, ok := accumulator.buffer.Load(event.AgentID); !ok {
			accumulator.buffer.Store(event.AgentID, []provider.Event{})
		}

		value, _ := accumulator.buffer.Load(event.AgentID)
		events := value.([]provider.Event)

		accumulator.buffer.Store(
			event.AgentID,
			append(events, event),
		)

		out <- event
	}
}

func (accumulator *Accumulator) String() string {
	var buffer strings.Builder

	accumulator.buffer.Range(func(key, value any) bool {
		events := value.([]provider.Event)

		for _, event := range events {
			buffer.WriteString(event.Content)
		}

		return true
	})

	return buffer.String()
}
