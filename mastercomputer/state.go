package mastercomputer

import "sync"

var stateManager *StateManager
var onceStateManager sync.Once

type StateManager struct {
	workers map[string]*Worker
	worker  *Worker
	lock    sync.Mutex
}

func NewStateManager() *StateManager {
	onceStateManager.Do(func() {
		stateManager = &StateManager{workers: make(map[string]*Worker)}
	})
	return stateManager
}

func (manager *StateManager) SetState(parameters map[string]any) *StateManager {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	workerID := parameters["worker_id"].(string)
	state := parameters["state"].(string)

	switch state {
	case "ready":
		manager.workers[workerID].state = WorkerStateReady
	case "working":
		manager.workers[workerID].state = WorkerStateWorking
	case "reviewing":
		manager.workers[workerID].state = WorkerStateReviewing
	case "done":
		manager.workers[workerID].state = WorkerStateDone
	}

	return manager
}

func (manager *StateManager) Start() string {
	return "state changed to " + manager.worker.State()
}
