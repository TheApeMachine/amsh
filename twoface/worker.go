package twoface

import (
	"context"
	"sync"
	"time"
)

// Worker processes jobs from the job channel.
type Worker struct {
	ID           int
	WorkerPool   chan chan Job
	JobChannel   chan Job
	ctx          context.Context
	wg           *sync.WaitGroup
	lastUse      time.Time
	lastDuration int64
	drain        bool
}

// NewWorker creates a new worker.
func NewWorker(ID int, workerPool chan chan Job, ctx context.Context) *Worker {
	return &Worker{
		ID:         ID,
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		ctx:        ctx,
		lastUse:    time.Now(),
		drain:      false,
	}
}

// Start the worker to be ready to accept jobs from the job queue.
func (worker *Worker) Read(p []byte) (n int, err error) {
	worker.wg.Add(1)

	go func() {
		for {
			worker.WorkerPool <- worker.JobChannel
		}
	}()

	worker.wg.Wait()
	return
}

func (worker *Worker) Write(p []byte) (n int, err error) {
	worker.wg.Add(1)

	go func() {
		for {
			worker.WorkerPool <- worker.JobChannel

			select {
			case job := <-worker.JobChannel:
				worker.lastUse = time.Now()
				job.Read(p)
			case <-worker.ctx.Done():
				return
			}
		}
	}()

	worker.wg.Wait()
	return
}

func (worker *Worker) Close() error {
	return nil
}
