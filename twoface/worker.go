package twoface

import (
	"context"
	"fmt"
)

type Worker struct {
	ID         int
	JobChannel chan Job
	WorkerPool chan chan Job
	quit       chan struct{}
}

func NewWorker(id int, workerPool chan chan Job) *Worker {
	return &Worker{
		ID:         id,
		JobChannel: make(chan Job),
		WorkerPool: workerPool,
		quit:       make(chan struct{}),
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			// Register the worker's job channel into the worker pool.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// Process the received job.
				if err := job.Process(context.Background()); err != nil {
					fmt.Printf("Worker %d: error processing job: %v\n", w.ID, err)
				}
			case <-w.quit:
				// Stop the worker.
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.quit)
}
