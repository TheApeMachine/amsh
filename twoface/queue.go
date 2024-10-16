package twoface

import (
	"errors"
	"sync"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

// Queue connects subscribers via channels.
type Queue struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber
}

// Subscriber describes an agent connected to the queue.
type Subscriber struct {
	ID     string
	inbox  chan *data.Artifact
	topics map[string]struct{}
}

// queueInstance ensures that the queue is a singleton.
var queueInstance *Queue
var once sync.Once

// NewQueue returns the singleton instance of the queue.
func NewQueue() *Queue {
	once.Do(func() {
		queueInstance = &Queue{
			subscribers: make(map[string]*Subscriber),
		}
	})
	return queueInstance
}

// Register adds a new subscriber to the queue.
func (q *Queue) Register(ID string) (chan *data.Artifact, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.subscribers[ID]; exists {
		return nil, errors.New("subscriber already exists")
	}

	inbox := make(chan *data.Artifact, 128)
	q.subscribers[ID] = &Subscriber{
		ID:     ID,
		inbox:  inbox,
		topics: make(map[string]struct{}),
	}

	return inbox, nil
}

// Publish sends a message to all relevant subscribers.
func (q *Queue) Publish(message *data.Artifact) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	publisher := message.Peek("origin")
	topic := message.Peek("scope")

	if publisher == "" || topic == "" {
		return errors.New("message missing origin or scope")
	}

	for _, subscriber := range q.subscribers {
		if topic == "broadcast" || subscriber.ID == topic {
			select {
			case subscriber.inbox <- message:
			default:
				errnie.Warn("message %s not delivered to %s", message.Peek("id"), subscriber.ID)
			}
		} else if _, subscribed := subscriber.topics[topic]; subscribed {
			select {
			case subscriber.inbox <- message:
			default:
				errnie.Warn("message %s not delivered to %s", message.Peek("id"), subscriber.ID)
			}
		}
	}

	return nil
}

// Unregister removes a subscriber from the queue.
func (q *Queue) Unregister(ID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return errors.New("subscriber does not exist")
	}

	close(subscriber.inbox)
	delete(q.subscribers, ID)
	return nil
}
