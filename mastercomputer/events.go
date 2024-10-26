package mastercomputer

import (
	"sync"
	"time"
)

// Event represents an event in the sequencer for tracking purposes.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	WorkerID  string    `json:"worker_id,omitempty"`
}

var events *Events
var onceEvents sync.Once

type Events struct {
	channel chan Event
}

func NewEvents() *Events {
	onceEvents.Do(func() {
		events = &Events{
			channel: make(chan Event, 64),
		}
	})

	return events
}

func (e *Events) Stream() <-chan Event {
	return e.channel
}
