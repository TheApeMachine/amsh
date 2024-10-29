package twoface

import (
	"context"
	"sync"
)

var (
	queueInstance *Queue
	onceQueue     sync.Once
)

/*
Queue is a simple pub/sub implementation that allows for topics to be created
on the fly and for subscribers to be added and removed dynamically.
*/
type Queue struct {
	ctx    context.Context
	cancel context.CancelFunc
}

/*
NewQueue instantiates a new queue, creating a context and cancel function
that can be used to signal the queue to stop processing new jobs.
It is an ambient context, so multiple calls to this function will return
the same instance.
*/
func NewQueue() *Queue {
	onceQueue.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		queueInstance = &Queue{
			ctx:    ctx,
			cancel: cancel,
		}
	})

	return queueInstance
}
