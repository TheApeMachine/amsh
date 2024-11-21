package provider

import (
	"strings"
	"sync"
)

type Accumulator struct {
	buffer *sync.Map
}

func NewAccumulator() *Accumulator {
	return &Accumulator{buffer: &sync.Map{}}
}

func (accumulator *Accumulator) Stream(
	in <-chan Event,
	sinks ...chan<- Event,
) {
	for event := range in {
		if _, ok := accumulator.buffer.Load(event.AgentID); !ok {
			accumulator.buffer.Store(event.AgentID, []Event{})
		}

		value, _ := accumulator.buffer.Load(event.AgentID)
		events := value.([]Event)

		accumulator.buffer.Store(
			event.AgentID,
			append(events, event),
		)

		for _, sink := range sinks {
			select {
			case sink <- event:
			default:
			}
		}
	}
}

func (accumulator *Accumulator) Collect(
	in <-chan Event,
) string {
	var buffer strings.Builder

	for event := range in {
		buffer.WriteString(event.Content)
	}

	return buffer.String()
}

func (accumulator *Accumulator) String() string {
	var buffer strings.Builder

	accumulator.buffer.Range(func(key, value any) bool {
		events := value.([]Event)

		for _, event := range events {
			buffer.WriteString(event.Content)
		}

		return true
	})

	return buffer.String()
}
