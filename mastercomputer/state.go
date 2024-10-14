package mastercomputer

import "github.com/theapemachine/amsh/errnie"

type WorkerState uint

const (
	// WorkerStateCreating is the starting state of the worker, indicating that it has not been initialized yet.
	WorkerStateCreating WorkerState = iota

	// WorkerStateInitializing indicates we are intializing, and nothing has gone wrong yet.
	WorkerStateInitializing

	// WorkerStateReady indicates that the worker is ready to take on work.
	WorkerStateReady

	// WorkerStateAcknowledged indicates the worker received a message and is sending ACK to the sender.
	WorkerStateAcknowledged

	// WorkerStateAccepted indicates the worker aceepted a workload and is now the owner of it.
	WorkerStateAccepted

	// WorkerStateRejected indicates the worker rejected a workload.
	WorkerStateRejected

	// WorkerStateBusy indicates the worker is currently actively performing work.
	WorkerStateBusy

	// WorkerStateWaiting indicates the worker is actively performing work, or about to, but waiting for additional input.
	WorkerStateWaiting

	// WorkerStateDone indicates the worker has completed the work it was previously busy with.
	WorkerStateDone

	// WorkerStateError indicates the worker has experienced an error.
	WorkerStateError

	// WorkerStateFinished indicates the worker has been deallocated and is shutting down.
	WorkerStateFinished

	// WorkerStateZombie indicates the worker has experienced a fatal error and is shutting down.
	WorkerStateZombie
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
}

/*
NewState returns a new state, provided that the worker is able to switch to that state.
If the requested state is not an allowed transition, the current state is returned.
The worker will immediately transition to the new state if it is allowed, so this
should only be used when the worker intends to transition to the new state.
*/
func (worker *Worker) NewState(state WorkerState) WorkerState {
	errnie.Trace()

	if worker.IsAllowed(state) {
		worker.State = state
	}

	return worker.State
}

/*
IsAllowed returns true if the worker can transition to the new state.
*/
func (worker *Worker) IsAllowed(state WorkerState) bool {
	errnie.Trace()

	if _, ok := transitions[worker.State]; !ok {
		return false
	}

	for _, allowed := range transitions[worker.State] {
		if allowed == state {
			return true
		}
	}

	return false
}
