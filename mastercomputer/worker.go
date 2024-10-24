package mastercomputer

import (
	"context"
	"math/rand/v2"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type WorkerState uint

const (
	WorkerStateUndecided WorkerState = iota
	WorkerStateDiscussing
	WorkerStateAgreed
	WorkerStateDisagreed
)

/*
Worker represents a blank agent that can adopt any role or workload that is assigned to it.
By setting the system prompt, user prompt, and assigning a toolset, the worker is flexible enough
to be the only agentic type. Finally, given that a worker is also a tool, workers that are assigned
the worker tool can make their own workers and delegate work.
*/
type Worker struct {
	parentCtx   context.Context
	executor    *Executor
	name        string
	role        string
	system      string
	user        string
	toolset     []openai.ChatCompletionToolParam
	temperature float64
	state       WorkerState
	discussion  *Executor
}

/*
NewWorker provides a minimal, uninitialized worker object. The buffer artifact is used to
initialize the worker's configuration, and an essential part of data transfer.
*/
func NewWorker(
	ctx context.Context,
	name string,
	toolset []openai.ChatCompletionToolParam,
	executor *Executor,
	role string,
) *Worker {
	errnie.Trace()

	return &Worker{
		parentCtx: ctx,
		name:      name,
		role:      role,
		toolset:   toolset,
		executor:  executor,
	}
}

/*
Initialize sets up the worker's context, queue registration, and initializes the memory.
This is the preparation phase for the worker to be ready to receive and process messages.
*/
func (worker *Worker) Initialize() *Worker {
	errnie.Trace()

	worker.temperature = utils.ToFixed(rand.Float64()*1.0, 1)
	return worker
}

func (worker *Worker) Start() {
	errnie.Trace()

	// Use the worker's history when calling Do
	if worker.state == WorkerStateDiscussing || worker.state == WorkerStateDisagreed || worker.discussion != nil {
		worker.discussion.Do(worker)
		return
	}

	worker.executor.Do(worker)
}
