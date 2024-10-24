package twoface

import (
	"context"
	"errors"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

/*
Queue is a simple pub/sub implementation that allows for topics to be created
on the fly and for subscribers to be added and removed dynamically.
*/
type Queue struct {
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	subscribers  map[string]*Subscriber
	PubCh        chan data.Artifact
	EventEmitter *EventEmitter
}

/*
Subscriber describes an agent connected to the queue.
*/
type Subscriber struct {
	ID     string
	inbox  chan data.Artifact
	topics map[string]struct{}
}

var queueInstance *Queue
var once sync.Once

/*
NewQueue returns the singleton instance of the queue.
*/
func NewQueue() *Queue {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		queueInstance = &Queue{
			ctx:          ctx,
			cancel:       cancel,
			subscribers:  make(map[string]*Subscriber),
			PubCh:        make(chan data.Artifact, 128),
			EventEmitter: NewEventEmitter(),
		}
		queueInstance.Start()
	})

	return queueInstance
}

/*
Start the queue to begin processing messages.
*/
func (q *Queue) Start() {
	go func() {
		for {
			select {
			case <-q.ctx.Done():
				return
			case message := <-q.PubCh:
				switch message.Peek("role") {
				case "register":
					q.Register(message.Peek("origin"))
				case "unregister":
					q.Unregister(message.Peek("origin"))
				case "subscribe":
					q.Subscribe(message.Peek("origin"), message.Peek("scope"))
				case "unsubscribe":
					q.Unsubscribe(message.Peek("origin"), message.Peek("scope"))
				default:
					q.Publish(message)
				}
			}
		}
	}()
}

/*
Stop the queue from processing messages.
*/
func (q *Queue) Stop() {
	q.cancel()
}

/*
GetTopics returns a list of all topics.
*/
func (q *Queue) GetTopics() []string {
	q.mu.RLock()
	defer q.mu.RUnlock()

	topicSet := make(map[string]struct{})
	for _, subscriber := range q.subscribers {
		for topic := range subscriber.topics {
			topicSet[topic] = struct{}{}
		}
	}

	topics := make([]string, 0, len(topicSet))
	for topic := range topicSet {
		topics = append(topics, topic)
	}

	return topics
}

/*
Register adds a new subscriber to the queue.
*/
func (q *Queue) Register(ID string) (chan data.Artifact, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.subscribers[ID]; exists {
		return nil, errors.New("subscriber already exists")
	}

	inbox := make(chan data.Artifact, 128)
	q.subscribers[ID] = &Subscriber{
		ID:     ID,
		inbox:  inbox,
		topics: make(map[string]struct{}),
	}

	return inbox, nil
}

// Subscribe adds a topic to a subscriber.
func (q *Queue) Subscribe(ID string, topic string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return errors.New("subscriber does not exist")
	}

	subscriber.topics[topic] = struct{}{}
	return nil
}

/*
Unsubscribe from a topic.
*/
func (q *Queue) Unsubscribe(ID string, topic string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return errnie.Error(errors.New("subscriber does not exist"))
	}

	delete(subscriber.topics, topic)
	return nil
}

// Publish sends a message to all relevant subscribers.
func (q *Queue) Publish(message data.Artifact) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	publisher := message.Peek("origin")
	topic := message.Peek("scope")

	if publisher == "" || topic == "" {
		spew.Dump(message)
		return errnie.Error(errors.New(
			"message not published, invalid origin or topic",
		))
	}

	errnie.Debug("Publishing message with ID: %s", message.Peek("id"))
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

	q.EmitEvent(EventTypeQueueUpdate, message)

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

// Add this method to the Queue struct
func (q *Queue) EmitEvent(eventType EventType, payload data.Artifact) {
	q.EventEmitter.Emit(Event{Type: eventType, Payload: payload})
}
