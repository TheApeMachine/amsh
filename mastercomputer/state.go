package mastercomputer

// WorkerState represents the state of a Worker.
type WorkerState uint

const (
	WorkerStateCreating WorkerState = iota
	WorkerStateInitializing
	WorkerStateReady
	WorkerStateAcknowledged
	WorkerStateAccepted
	WorkerStateRejected
	WorkerStateBusy
	WorkerStateWaiting
	WorkerStateDone
	WorkerStateError
	WorkerStateFinished
	WorkerStateZombie
	WorkerStateNotOK
)

var transitions = map[WorkerState][]WorkerState{
	WorkerStateCreating:     {WorkerStateInitializing},
	WorkerStateInitializing: {WorkerStateReady},
	WorkerStateReady:        {WorkerStateAcknowledged, WorkerStateAccepted},
	WorkerStateAcknowledged: {WorkerStateAccepted},
	WorkerStateAccepted:     {WorkerStateBusy, WorkerStateWaiting},
	WorkerStateBusy:         {WorkerStateWaiting, WorkerStateDone},
	WorkerStateWaiting:      {WorkerStateBusy, WorkerStateDone},
	WorkerStateDone:         {WorkerStateFinished},
	WorkerStateError:        {WorkerStateFinished},
	WorkerStateFinished:     {},
	WorkerStateZombie:       {},
	WorkerStateNotOK:        {},
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

// String returns the string representation of the worker state.
func (state WorkerState) String() string {
	switch state {
	case WorkerStateCreating:
		return "creating"
	case WorkerStateInitializing:
		return "initializing"
	case WorkerStateReady:
		return "ready"
	case WorkerStateAcknowledged:
		return "acknowledged"
	case WorkerStateAccepted:
		return "accepted"
	case WorkerStateRejected:
		return "rejected"
	case WorkerStateBusy:
		return "busy"
	case WorkerStateWaiting:
		return "waiting"
	case WorkerStateDone:
		return "done"
	case WorkerStateError:
		return "error"
	case WorkerStateFinished:
		return "finished"
	case WorkerStateZombie:
		return "zombie"
	}

	return ""
}
