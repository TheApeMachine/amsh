package mastercomputer

import (
	"sync"

	"github.com/theapemachine/amsh/ai"
)

var managerInstance *Manager
var managerOnce sync.Once

// Manager manages all active workers.
type Manager struct {
	wg      sync.WaitGroup
	workers map[string]*Worker
	mu      sync.Mutex
	memory  *ai.Memory
}

// NewWorkerManager creates a new WorkerManager.
func NewManager() *Manager {
	managerOnce.Do(func() {
		managerInstance = &Manager{
			workers: make(map[string]*Worker),
			memory:  ai.NewMemory("hive"),
		}
	})
	return managerInstance
}

// AddWorker adds a new worker to the manager and increments the WaitGroup.
func (manager *Manager) AddWorker(worker *Worker) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.workers[worker.buffer.Peek("origin")] = worker
	manager.wg.Add(1)
}

func (manager *Manager) GetWorker(workerID string) *Worker {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	return manager.workers[workerID]
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
