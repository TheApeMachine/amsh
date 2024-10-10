package core

import (
	"sync"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

/*
Queue is the method to send and receive messages between types.
It uses a map to manage topic-based subscriptions and a mutex for thread safety.
The message type is data.Artifact.
*/
type Queue struct {
	topicChans map[string][]chan *data.Artifact
	mutex      sync.RWMutex
}

// NewQueue initializes a new Queue.
func NewQueue() *Queue {
	return &Queue{
		topicChans: make(map[string][]chan *data.Artifact),
	}
}

// Subscribe allows a subscriber to listen to a specific topic.
// It returns a channel to receive artifacts related to the topic.
func (q *Queue) Subscribe(topic string) <-chan *data.Artifact {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	ch := make(chan *data.Artifact, 100) // Buffered channel to prevent blocking
	q.topicChans[topic] = append(q.topicChans[topic], ch)
	return ch
}

// Unsubscribe removes a subscriber's channel from a specific topic.
func (q *Queue) Unsubscribe(topic string, ch <-chan *data.Artifact) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	subscribers, ok := q.topicChans[topic]
	if !ok {
		return
	}

	for i, subscriber := range subscribers {
		if subscriber == ch {
			// Remove the subscriber from the slice
			q.topicChans[topic] = append(subscribers[:i], subscribers[i+1:]...)
			close(subscriber)
			break
		}
	}

	// Clean up the topic if no subscribers remain
	if len(q.topicChans[topic]) == 0 {
		delete(q.topicChans, topic)
	}
}

// Publish sends an artifact to all subscribers of the given topic.
func (q *Queue) Publish(topic string, artifact *data.Artifact) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	for _, ch := range q.topicChans[topic] {
		select {
		case ch <- artifact:
		default:
			// Handle full channel, e.g., skip or log
			errnie.Warn("Publish artifact to topic '%s' failed: channel is full", topic)
		}
	}
}

// Close gracefully closes all subscriber channels.
func (q *Queue) Close() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for topic, subscribers := range q.topicChans {
		for _, ch := range subscribers {
			close(ch)
		}
		delete(q.topicChans, topic)
	}
}
