package twoface

import (
	"context"
	"sync"
)

var (
	globalPool *Pool
	once       sync.Once
)

func SetGlobalPool(pool *Pool) {
	once.Do(func() {
		globalPool = pool
	})
}

func GetGlobalPool() *Pool {
	return globalPool
}

type Pool struct {
	ctx        context.Context
	cancel     context.CancelFunc
	JobQueue   chan Job
	WorkerPool chan chan Job
	workers    []*Worker
	mu         sync.Mutex
}

func NewPool(ctx context.Context, initialWorkers int) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	pool := &Pool{
		ctx:        ctx,
		cancel:     cancel,
		JobQueue:   make(chan Job),
		WorkerPool: make(chan chan Job),
		workers:    make([]*Worker, 0),
	}

	for i := 0; i < initialWorkers; i++ {
		pool.addWorker()
	}

	go pool.dispatch()

	return pool
}

func (p *Pool) addWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	worker := NewWorker(len(p.workers), p.WorkerPool)
	worker.Start()
	p.workers = append(p.workers, worker)
}

func (p *Pool) removeWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.workers) == 0 {
		return
	}

	worker := p.workers[len(p.workers)-1]
	worker.Stop()
	p.workers = p.workers[:len(p.workers)-1]
}

func (p *Pool) dispatch() {
	for {
		select {
		case job := <-p.JobQueue:
			go func(job Job) {
				// Obtain an available worker's job channel.
				jobChannel := <-p.WorkerPool
				// Send the job to the worker.
				jobChannel <- job
			}(job)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pool) Submit(job Job) {
	select {
	case p.JobQueue <- job:
	case <-p.ctx.Done():
		// Pool is shutting down.
	}
}

func (p *Pool) Shutdown() {
	p.cancel()
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, worker := range p.workers {
		worker.Stop()
	}
	p.workers = nil
}
