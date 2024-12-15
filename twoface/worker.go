package twoface

import (
	"time"

	"github.com/theapemachine/errnie"
)

type PoolWorker interface {
	Start() PoolWorker
	Drain()
	LastUse() time.Time
	LastDuration() int64
}

type Worker struct {
	ID           int
	WorkerPool   chan chan Job
	JobChannel   chan Job
	queue        *Queue
	lastUse      time.Time
	lastDuration int64
	drain        bool
}

func NewWorker(
	ID int,
	workerPool chan chan Job,
) *Worker {
	return &Worker{
		ID:         ID,
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		queue:      NewQueue(),
		lastUse:    time.Now(),
		drain:      false,
	}
}

/*
Start the worker to be ready to accept jobs from the job queue.
*/
func (worker *Worker) Start() PoolWorker {
	go func() {
		defer close(worker.JobChannel)

		for {
			// Return the job channel to the worker pool.
			worker.WorkerPool <- worker.JobChannel

			// Pick up a new job if available.
			job := <-worker.JobChannel

			// Keep track of the time before the work starts, with a
			// secondary benefit of helping to determine if the worker
			// is idle for a significant amount of time later on.
			worker.lastUse = time.Now()

			// Execute the job.
			job.Do()

			// Store the duration of the job load so it can later be used to
			// determine if the worker pool is overloaded.
			worker.lastDuration = time.Since(worker.lastUse).Nanoseconds()

			// This worker is about to get retired in a pool schrink.
			if worker.drain {
				return
			}
		}
	}()

	return worker
}

/*
Drain the worker, which means it will finish its current job first
before it will stop.
*/
func (worker *Worker) Drain() {
	errnie.Trace("draining worker %d", worker.ID)
	worker.drain = true
}

/*
LastUse returns the time the worker was last used.
*/
func (worker *Worker) LastUse() time.Time {
	return worker.lastUse
}

/*
LastDuration returns the duration of the last job the worker executed.
*/
func (worker *Worker) LastDuration() int64 {
	return worker.lastDuration
}
