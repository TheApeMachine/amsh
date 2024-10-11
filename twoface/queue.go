package twoface

import (
	"errors"
	"sync"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

var queueCache *Queue

type Message struct {
	Topic string
	Data  data.Artifact
}

type Subscriber struct {
	ID       string
	Channels map[string]chan data.Artifact
}

type Queue struct {
	mu           sync.RWMutex
	subscribers  map[string]*Subscriber
	globalTopics map[string][]*Subscriber
}

func NewQueue() *Queue {
	if queueCache == nil {
		queueCache = &Queue{
			subscribers:  make(map[string]*Subscriber),
			globalTopics: make(map[string][]*Subscriber),
		}
	}

	return queueCache
}

// Register a new subscriber
func (q *Queue) Register(ID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.subscribers[ID]; !exists {
		q.subscribers[ID] = &Subscriber{
			ID:       ID,
			Channels: make(map[string]chan data.Artifact),
		}
	}
}

// Subscribe to a topic
func (q *Queue) Subscribe(ID, topic string) (chan data.Artifact, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return nil, errnie.Error(errors.New("Subscriber does not exist"))
	}

	ch := make(chan data.Artifact, 100) // Buffered channel
	subscriber.Channels[topic] = ch

	q.globalTopics[topic] = append(q.globalTopics[topic], subscriber)

	return ch, nil
}

// Publish a message to a topic
func (q *Queue) Publish(topic string, message data.Artifact) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	subscribers, exists := q.globalTopics[topic]
	if !exists {
		return
	}

	for _, subscriber := range subscribers {
		ch, ok := subscriber.Channels[topic]
		if ok {
			select {
			case ch <- message:
				// Message sent
			default:
				errnie.Warn("Dropping message for subscriber %s on topic %s: channel full", subscriber.ID, topic)
			}
		}
	}
}

// Unsubscribe from a topic
func (q *Queue) Unsubscribe(ID, topic string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return
	}

	if ch, ok := subscriber.Channels[topic]; ok {
		close(ch)
		delete(subscriber.Channels, topic)
	}

	// Remove subscriber from globalTopics
	subscribers := q.globalTopics[topic]
	for i, sub := range subscribers {
		if sub.ID == ID {
			q.globalTopics[topic] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
}

// Unregister a subscriber
func (q *Queue) Unregister(ID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return
	}

	// Close all channels
	for _, ch := range subscriber.Channels {
		close(ch)
	}

	// Remove from global topics
	for topic := range subscriber.Channels {
		subscribers := q.globalTopics[topic]
		for i, sub := range subscribers {
			if sub.ID == ID {
				q.globalTopics[topic] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}

	delete(q.subscribers, ID)
}
