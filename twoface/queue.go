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
	subscriptions []Subscriber
	I             chan data.Artifact
}

func NewQueue() *Queue {
	if queueCache == nil {
		queueCache = &Queue{
			subscriptions: make([]Subscriber, 0),
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
			default:
				// noop
			}
		}
	}()
}

func (q *Queue) Register(ID string) chan data.Artifact {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.subscriptions = append(q.subscriptions, Subscriber{
		ID:     ID,
		O:      make(chan data.Artifact, 100), // Buffered channel to prevent blocking
		topics: make(map[string]chan data.Artifact),
	})

	return q.subscriptions[len(q.subscriptions)-1].O
}

func (q *Queue) Subscribe(ID, topic string) chan data.Artifact {
	q.mu.Lock()
	defer q.mu.Unlock()

	ch := make(chan data.Artifact, 100) // Buffered channel for topic subscriptions

	for i, subscriber := range q.subscriptions {
		if subscriber.ID == ID {
			q.subscriptions[i].topics[topic] = ch
			break
		}
	}

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
			select {
			case subscriber.O <- message:
				// Message sent successfully
			default:
				errnie.Warn("Dropping message for subscriber %s: channel full", subscriber.ID)
			}
		}
	}

	if role == "topic" {
		for _, subscriber := range q.subscriptions {
			if ch, ok := subscriber.topics[scope]; ok {
				select {
				case ch <- message:
					// Message sent successfully
				default:
					errnie.Warn("Dropping message for subscriber %s on topic %s: channel full", subscriber.ID, scope)
				}
			}
		}
	}

	if role == "report" || role == "ACK" {
		for _, subscriber := range q.subscriptions {
			if subscriber.ID == scope {
				select {
				case subscriber.O <- message:
					// Message sent successfully
				default:
					errnie.Warn("Dropping message for subscriber %s: channel full", subscriber.ID)
				}
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
