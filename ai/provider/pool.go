package provider

import (
	"context"
	"sync"

	"github.com/charmbracelet/log"
)

type ProviderPool struct {
	jobs    chan *GenerationJob
	workers chan chan *GenerationJob
	mu      sync.Mutex // Controls access to the providers
}

type Worker struct {
	pool      *ProviderPool
	jobChan   chan *GenerationJob
	provider  Provider
	failCount int
}

func NewWorker(pool *ProviderPool, provider Provider) *Worker {
	return &Worker{
		pool:     pool,
		jobChan:  make(chan *GenerationJob),
		provider: provider,
	}
}

func (worker *Worker) Start() {
	go func() {
		for {
			worker.pool.workers <- worker.jobChan
			job := <-worker.jobChan
			for event := range worker.provider.Generate(job.ctx, job.params, job.messages) {
				if event.Type == EventDone {
					close(job.resultChan)
					return
				}

				if event.Type == EventError {
					worker.failCount++

					if worker.failCount > 3 {
						close(job.resultChan)
						return
					}

					break
				}

				job.resultChan <- event
			}
		}
	}()
}

type GenerationJob struct {
	ctx        context.Context
	params     GenerationParams
	messages   []Message
	resultChan chan Event
}

func NewProviderPool(providers []Provider) *ProviderPool {
	pool := &ProviderPool{
		jobs:    make(chan *GenerationJob, 4),
		workers: make(chan chan *GenerationJob, 4),
	}

	// Initialize provider statuses
	for _, provider := range providers {
		worker := NewWorker(pool, provider)
		go worker.Start()
	}

	// Start the dispatcher
	go pool.dispatch()

	return pool
}

// Schedule job to be processed by the next available provider
func (pool *ProviderPool) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	resultChan := make(chan Event)
	pool.jobs <- &GenerationJob{
		ctx:        ctx,
		params:     params,
		messages:   messages,
		resultChan: resultChan,
	}

	log.Info("job scheduled")
	return resultChan
}

/*
dispatch is the main loop for the dispatcher, which takes an available worker
from the worker queue and assigns it a job from the jobs queue.
*/
func (pool *ProviderPool) dispatch() {
	// Make sure that we cleanly close the channels if our dispatcher
	// returns for whatever reason.
	defer close(pool.jobs)
	defer close(pool.workers)

	for {
		select {
		case job := <-pool.jobs:
			// A new job was received from the jobs queue, get the first available
			// worker from the pool once ready.
			jobChannel := <-pool.workers
			// Then send the job to the worker for processing.
			jobChannel <- job
		}
	}
}

func (pool *ProviderPool) Configure(config map[string]any) {
	// TODO: Implement configuration
}
