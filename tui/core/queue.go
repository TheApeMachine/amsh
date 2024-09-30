package core

import (
	"sync" // Added for mutex

	"github.com/theapemachine/amsh/data"
)

/*
Queue is the method to send and receive messages between types.
It uses a ring buffer to store messages and a mutex to synchronize access.
The message type is data.Artifact.
*/
type Queue struct {
	chans      []chan data.Artifact
	topicChans map[string][]chan *data.Artifact // New field to manage topics
	mutex      sync.RWMutex                     // New field for thread safety
}

func NewQueue(size int) *Queue {
	return &Queue{
		chans:      make([]chan data.Artifact, size),
		topicChans: make(map[string][]chan *data.Artifact), // Initialize topic map
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
		}
	}
}

// Subscribe returns a channel to receive artifacts from the given topic.
func (q *Queue) Subscribe(topic string) <-chan *data.Artifact {
	ch := make(chan *data.Artifact, 100) // Buffer size as needed
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.topicChans[topic] = append(q.topicChans[topic], ch)
	return ch
}

// Unsubscribe removes a subscriber from a topic.
func (q *Queue) Unsubscribe(topic string, ch <-chan *data.Artifact) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	chans := q.topicChans[topic]
	for i, c := range chans {
		if c == ch {
			q.topicChans[topic] = append(chans[:i], chans[i+1:]...)
			close(c)
			break
		}
	}
}
