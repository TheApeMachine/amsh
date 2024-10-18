package mastercomputer

import (
	"sync"
)

// Manager manages all active workers.
type Manager struct {
	wg      sync.WaitGroup
	workers map[string]*Worker
	mu      sync.Mutex
}

// NewWorkerManager creates a new WorkerManager.
func NewManager() *Manager {
	return &Manager{
		workers: make(map[string]*Worker),
	}
}

// AddWorker adds a new worker to the manager and increments the WaitGroup.
func (manager *Manager) AddWorker(worker *Worker) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.workers[worker.buffer.Peek("id")] = worker
	manager.wg.Add(1)
}

// RemoveWorker removes a worker from the manager and decrements the WaitGroup.
func (manager *Manager) RemoveWorker(workerID string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	delete(manager.workers, workerID)
	manager.wg.Done()
}

// Wait waits for all workers to finish.
func (manager *Manager) Wait() {
	manager.wg.Wait()
}
