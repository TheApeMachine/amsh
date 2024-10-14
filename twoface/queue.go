package twoface

import (
	"errors"
	"sync"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

/*
queueCache ensures that the queue is an ambient context and everybody
is connected on the same channels.
*/
var queueCache *Queue

/*
Subscriber describes a Worker or Agent that is connected to the queue
and listening to specific channels.
*/
type Subscriber struct {
	ID     string
	inbox  chan data.Artifact
	topics []string
}

/*
Queue connects subscribers via channels. These channels can be either
broadcast channels, which are public, topics, which are meant for
specific groups, or private channels, which are for specific agents.
*/
type Queue struct {
	mu          sync.RWMutex
	subscribers map[string]*Subscriber
}

/*
NewQueue instantiates the queue, or returns the queue from the cache
if it was previously created.
*/
func NewQueue() *Queue {
	errnie.Trace()

	if queueCache == nil {
		queueCache = &Queue{
			subscribers: make(map[string]*Subscriber),
		}
	}

	return queueCache
}

/*
Register should be called by all newly created agents to patch them into
the communication network.
*/
func (q *Queue) Register(ID string) (chan data.Artifact, error) {
	errnie.Info("registering %s", ID)

	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.subscribers[ID]; !exists {
		inbox := make(chan data.Artifact, 128)
		q.subscribers[ID] = &Subscriber{
			ID:     ID,
			inbox:  inbox,
			topics: make([]string, 0),
		}

		errnie.Info("registered %s", ID)
	}

	return q.subscribers[ID].inbox, nil
}

// Subscribe to a topic
func (q *Queue) Subscribe(ID, topic string) (err error) {
	errnie.Trace()

	q.mu.Lock()
	defer q.mu.Unlock()

	errnie.Info("subscribing %s to %s", ID, topic)

	// Check if the subscriber exists.
	subscriber, exists := q.subscribers[ID]
	if !exists {
		return errnie.Error(errors.New("Subscriber does not exist"))
	}

	// Check if the subscriber is already subscribed to the topic.
	for _, t := range subscriber.topics {
		if t == topic {
			return errnie.Error(errors.New("subscriber is already subscribed to topic"))
		}
	}

	// Add the topic to the subscriber.
	subscriber.topics = append(subscriber.topics, topic)
	errnie.Info("subscribed %s to %s", ID, topic)

	return nil
}

/*
Publish a message onto the queue, provided all the necessary conditions are met.
*/
func (q *Queue) Publish(message data.Artifact) (err error) {
	errnie.Trace()

	q.mu.RLock()
	defer q.mu.RUnlock()

	// Check the origin of the message to see if the subscriber
	// exists, and is subscribed to the topic.
	publisher := message.Peek("origin")
	topic := message.Peek("scope")

	// Check if the message has an origin and a scope, or we cannot proceed.
	if publisher == "" || topic == "" {
		return errnie.Error(errors.New("message is missing origin or topic"))
	}

	// Check if the publisher is registered
	if _, exists := q.subscribers[publisher]; !exists {
		return errnie.Error(errors.New("publisher is not registered"))
	}

	// Check if the scope is broadcast, in which case we send to all subscribers.
	if topic == "broadcast" {
		for _, subscriber := range q.subscribers {
			subscriber.inbox <- message
		}

		return nil
	}

	// Check if the scope matches a subscriber ID, in which case
	// we are dealing with a private message.
	if _, exists := q.subscribers[topic]; exists {
		q.subscribers[topic].inbox <- message
		return nil
	}

	// Check if the publisher is subscribed to the topic.
	subscribed := false

	for _, t := range q.subscribers[publisher].topics {
		if t == topic {
			subscribed = true
			break
		}
	}

	if !subscribed {
		return errnie.Error(errors.New("publisher is not subscribed to topic"))
	}

	// Otherwise, we send to all subscribers of the topic.
	for _, subscriber := range q.subscribers {
		for _, t := range subscriber.topics {
			if t == topic {
				subscriber.inbox <- message
			}
		}
	}

	return nil
}

/*
Unsubscribe from a topic if an agent no longer needs to respond to updates on
the channel.
*/
func (q *Queue) Unsubscribe(ID, topic string) (err error) {
	errnie.Trace()

	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return errnie.Error(errors.New("subscriber does not exist"))
	}

	for i, t := range subscriber.topics {
		if t == topic {
			// Remove the topic from the subscriber.
			subscriber.topics = append(subscriber.topics[:i], subscriber.topics[i+1:]...)
			break
		}
	}

	return nil
}

/*
Unregister from the queue, which should be called in an agent's life-cycle exit stage.
*/
func (q *Queue) Unregister(ID string) (err error) {
	errnie.Trace()

	q.mu.Lock()
	defer q.mu.Unlock()

	subscriber, exists := q.subscribers[ID]
	if !exists {
		return errnie.Error(errors.New("subscriber does not exist"))
	}

	// Close the subscriber's inbox.
	close(subscriber.inbox)

	// Delete the subscriber from the queue.
	delete(q.subscribers, ID)

	return nil
}
