package twoface

import (
	"context"
	"sync"
)

type Process interface {
	Ctx() context.Context
	ID() string
	Name() string
	Start()
	Stop()
	Manager() *WorkerManager
}

// WorkerManager manages all active workers.
type WorkerManager struct {
	wg      sync.WaitGroup
	workers map[string]Process
	mu      sync.Mutex
}

// NewWorkerManager creates a new WorkerManager.
func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		workers: make(map[string]Process),
	}
}

// AddWorker adds a new worker to the manager and increments the WaitGroup.
func (wm *WorkerManager) AddWorker(worker Process) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workers[worker.ID()] = worker
	wm.wg.Add(1)
}

// RemoveWorker removes a worker from the manager and decrements the WaitGroup.
func (wm *WorkerManager) RemoveWorker(workerID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.workers, workerID)
	wm.wg.Done()
}

// Wait waits for all workers to finish.
func (wm *WorkerManager) Wait() {
	wm.wg.Wait()
}
