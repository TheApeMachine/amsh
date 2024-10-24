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
	WorkerStateWorking
	WorkerStateDiscussing
	WorkerStateAgreed
	WorkerStateDisagreed
	WorkerStateDone
)

// Worker represents a flexible agent capable of handling different roles and workloads.
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
	buffer      *ConversationBuffer // Scoped buffer for this worker's context
}

// NewWorker creates a new worker with scoped context buffer.
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
		buffer:    NewConversationBuffer(name), // Assign scoped conversation buffer
	}
}

// Initialize sets up the worker's context, toolset, and memory.
func (worker *Worker) Initialize() *Worker {
	errnie.Trace()
	worker.temperature = utils.ToFixed(rand.Float64()*1.0, 1)
	return worker
}

// Start begins the worker's execution phase.
func (worker *Worker) Start() {
	errnie.Trace()
	worker.executor.Do(worker)
}
