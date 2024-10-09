package twoface

import (
	"context"
	"sync"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Subscriber struct {
	ID     string
	O      chan data.Artifact
	topics map[string]chan data.Artifact
}

var queueCache *Queue

type Queue struct {
	mu            sync.Mutex
	subscriptions map[string]Subscriber
	I             chan data.Artifact
}

func NewQueue() *Queue {
	if queueCache == nil {
		queueCache = &Queue{
			subscriptions: make(map[string]Subscriber),
			I:             make(chan data.Artifact, 128),
		}
	}

	return queueCache
}

func (q *Queue) Run(ctx context.Context) {
	go func() {
		defer close(q.I)

		for {
			select {
			case message := <-q.I:
				q.Publish(message)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (q *Queue) Register(ID string) chan data.Artifact {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.subscriptions[ID] = Subscriber{
		ID:     ID,
		O:      make(chan data.Artifact),
		topics: make(map[string]chan data.Artifact),
	}

	return q.subscriptions[ID].O
}

func (q *Queue) Subscribe(ID, topic string) chan data.Artifact {
	q.mu.Lock()
	defer q.mu.Unlock()

	ch := make(chan data.Artifact)
	q.subscriptions[ID].topics[topic] = ch
	return ch
}

func (q *Queue) Publish(message data.Artifact) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var (
		role  string
		scope string
		err   error
	)

	if role, err = message.Role(); errnie.Error(err) != nil {
		return
	}

	if scope, err = message.Scope(); errnie.Error(err) != nil {
		return
	}

	if scope == "broadcast" {
		for _, subscriber := range q.subscriptions {
			subscriber.O <- message
		}
	}

	if role == "topic" {
		// Loop over all the subscribers and send the message to each one who is subscribed to the topic.
		for _, subscriber := range q.subscriptions {
			if ch, ok := subscriber.topics[scope]; ok {
				ch <- message
			}
		}
	}

	if role == "report" || role == "ACK" {
		for _, subscriber := range q.subscriptions {
			if subscriber.ID == scope {
				subscriber.O <- message
			}
		}
	}
}

func (q *Queue) Close(topic string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, subscriber := range q.subscriptions {
		if ch, ok := subscriber.topics[topic]; ok {
			close(ch)
		}
	}
}
