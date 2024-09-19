package twoface

import (
	"context"
	"sync"
)

/*
Pool is a set of Worker types, each running their own (pre-warmed) goroutine.
Any object that implements the Job interface is able to schedule work on the
worker pool, which keeps the amount of goroutines in check, while still being
able to benefit from high concurrency in all kinds of scenarios.
*/
type Pool struct {
	ctx        context.Context
	cancel     context.CancelFunc
	workerPool chan chan Job
	jobQueue   chan Job
	workers    []*Worker
	wg         *sync.WaitGroup
}

/*
NewPool instantiates a worker pool with a given number of workers, taking in a
context for cleanly canceling all of the sub-processes it starts.
*/
func NewPool(ctx context.Context, numWorkers int) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}

	pool := &Pool{
		ctx:        ctx,
		cancel:     cancel,
		workerPool: make(chan chan Job, numWorkers),
		jobQueue:   make(chan Job),
		workers:    make([]*Worker, 0, numWorkers),
		wg:         wg,
	}

	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(i, pool.workerPool, ctx)
		worker.Start()
		pool.workers = append(pool.workers, worker)
	}

	go pool.dispatch()

	return pool
}

/*
Size returns the current size of the pool by counting the currently active workers.
*/
func (pool *Pool) Size() int {
	return len(pool.workers)
}

/*
Read is the entry point for new jobs that want to be scheduled onto the worker pool.
*/
func (pool *Pool) Read(p []byte) (n int, err error) {
	return 0, nil
}

/*
Write is the entry point for new jobs that want to be scheduled onto the worker pool.
*/
func (pool *Pool) Write(p []byte) (n int, err error) {
	pool.wg.Add(1)
	pool.jobQueue <- job
}

/*
Shutdown gracefully shuts down the pool and waits for all workers to complete.
*/
func (pool *Pool) Close() (err error) {
	pool.cancel()
	pool.wg.Wait()
	close(pool.jobQueue)
	close(pool.workerPool)

	for _, worker := range pool.workers {
		worker.Close()
	}

	return
}

func (pool *Pool) dispatch() {
	for {
		select {
		case job := <-pool.jobQueue:
			// A new job was received from the jobs queue, get the first available worker from the pool once ready.
			jobChannel := <-pool.workerPool
			// Send the job to the worker for processing.
			jobChannel <- job
		case <-pool.ctx.Done():
			return
		}
	}
}
