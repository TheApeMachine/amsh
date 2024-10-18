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

	errnie.Info("registering subscriber %s", ID)

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

// Subscribe adds a topic to a subscriber.
func (q *Queue) Subscribe(ID string, topic string) error {
	errnie.Info("subscribing %s to %s", ID, topic)

	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return errors.New("subscriber does not exist")
	}

	subscriber.topics[topic] = struct{}{}
	return nil
}

// Publish sends a message to all relevant subscribers.
func (q *Queue) Publish(message *data.Artifact) error {
	go func() {
		q.mu.RLock()
		defer q.mu.RUnlock()

		publisher := message.Peek("origin")
		topic := message.Peek("scope")

		if publisher == "" || topic == "" {
			errnie.Warn("message %s not published, invalid origin or topic", message.Peek("id"))
			return
		}

		sent := false
		for _, subscriber := range q.subscribers {
			if topic == "broadcast" || subscriber.ID == topic {
				sent = true
				select {
				case subscriber.inbox <- message:
				default:
					errnie.Warn("message %s not delivered to %s", message.Peek("id"), subscriber.ID)
				}
			} else if _, subscribed := subscriber.topics[topic]; subscribed {
				sent = true
				select {
				case subscriber.inbox <- message:
				default:
					errnie.Warn("message %s not delivered to %s", message.Peek("id"), subscriber.ID)
				}
			}
		}

		if !sent {
			// Create the new topic and subscribe the sender
			q.mu.RUnlock()
			q.mu.Lock()
			q.Subscribe(publisher, topic)
			q.mu.Unlock()
			q.mu.RLock()

			// Send a broadcast message about the new topic
			broadcastMsg := data.New(publisher, "system", "broadcast", []byte("New topic created: "+topic))
			q.Publish(broadcastMsg)

			// Publish the original message to the new topic
			q.Publish(message)
		}

		errnie.Info("%s -[%s]-> %s", message.Peek("origin"), message.Peek("role"), message.Peek("scope"))
		errnie.Note("[PAYLOAD]\n%s\n[/PAYLOAD]", message.Peek("payload"))
	}()

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
