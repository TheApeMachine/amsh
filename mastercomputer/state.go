package mastercomputer

// WorkerState represents the state of a Worker.
type WorkerState uint

const (
	WorkerStateUndefined WorkerState = iota
	WorkerStateCreating
	WorkerStateInitializing
	WorkerStateReady
	WorkerStateBusy
	WorkerStateError
	WorkerStateZombie
)

var transitions = map[WorkerState][]WorkerState{
	WorkerStateCreating:     {WorkerStateInitializing},
	WorkerStateInitializing: {WorkerStateReady},
	WorkerStateReady:        {WorkerStateBusy},
	WorkerStateBusy:         {WorkerStateReady},
	WorkerStateError:        {WorkerStateZombie},
	WorkerStateZombie:       {},
}

// NewState transitions the worker to a new state if allowed.
func (worker *Worker) NewState(state WorkerState) WorkerState {
	if worker.IsAllowed(state) {
		worker.state = state
	}
	return worker.state
}

// IsAllowed checks if the worker can transition to the new state.
func (worker *Worker) IsAllowed(state WorkerState) bool {
	if allowedStates, ok := transitions[worker.state]; ok {
		for _, allowed := range allowedStates {
			if allowed == state {
				return true
			}
		}
	}
	return false
}

func (worker *Worker) StateByKey(key string) WorkerState {
	switch key {
	case "creating":
		return WorkerStateCreating
	case "initializing":
		return WorkerStateInitializing
	case "busy":
		return WorkerStateBusy
	case "ready":
		return WorkerStateReady
	case "error":
		return WorkerStateError
	case "zombie":
		return WorkerStateZombie
	}

	return WorkerState(0)
}

// String returns the string representation of the worker state.
func (state WorkerState) String() string {
	switch state {
	case WorkerStateCreating:
		return "creating"
	case WorkerStateInitializing:
		return "initializing"
	case WorkerStateReady:
		return "ready"
	case WorkerStateBusy:
		return "busy"
	case WorkerStateError:
		return "error"
	case WorkerStateZombie:
		return "zombie"
	}

	return ""
}
