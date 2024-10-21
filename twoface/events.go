package twoface

import (
	"sync"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type EventType string

const (
	EventTypeQueueUpdate    EventType = "queue_update"
	EventTypeExecutorUpdate EventType = "executor_update"
)

type Event struct {
	Type    EventType
	Payload data.Artifact
}

type EventEmitter struct {
	mu        sync.RWMutex
	listeners map[chan Event]struct{}
}

var eventEmitterInstance *EventEmitter
var eventEmitterOnce sync.Once

func NewEventEmitter() *EventEmitter {
	eventEmitterOnce.Do(func() {
		eventEmitterInstance = &EventEmitter{
			listeners: make(map[chan Event]struct{}),
		}
	})
	return eventEmitterInstance
}

func (ee *EventEmitter) Subscribe() chan Event {
	ee.mu.Lock()
	defer ee.mu.Unlock()
	ch := make(chan Event, 100)
	ee.listeners[ch] = struct{}{}
	return ch
}

func (ee *EventEmitter) Unsubscribe(ch chan Event) {
	ee.mu.Lock()
	defer ee.mu.Unlock()
	delete(ee.listeners, ch)
	close(ch)
}

func (ee *EventEmitter) Emit(event Event) {
	if event.Payload.Peek("payload") == "" {
		errnie.Info("Skipping empty event of type: %s", event.Type)
		return
	}

	ee.mu.RLock()
	defer ee.mu.RUnlock()
	for ch := range ee.listeners {
		select {
		case ch <- event:
			errnie.Info("Emitted event of type: %s", event.Type)
		default:
			errnie.Warn("Channel full, skipped event of type: %s", event.Type)
		}
	}
}
