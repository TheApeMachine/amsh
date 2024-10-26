package mastercomputer

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
)

// WorkerState represents the current state of a worker.
type WorkerState uint

const (
	WorkerStateCreating WorkerState = iota
	WorkerStateReady
	WorkerStateWorking
	WorkerStateWaiting
	WorkerStateReviewing
	WorkerStateDone
)

// Worker represents a flexible agent capable of handling different roles and workloads.
type Worker struct {
	ctx     context.Context
	cancel  context.CancelFunc
	task    chan *Task
	name    string
	role    string
	toolset *Toolset
	state   WorkerState
	buffer  *Conversation
	events  *Events
}

// NewWorker creates a new worker with scoped context buffer.
func NewWorker(
	parameters map[string]any,
) *Worker {
	errnie.Trace()
	ctx, cancel := context.WithCancel(context.Background())
	name, nameOk := parameters["name"].(string)
	role, roleOk := parameters["role"].(string)

	if !nameOk || !roleOk {
		cancel()
		return nil
	}

	return &Worker{
		ctx:     ctx,
		cancel:  cancel,
		task:    make(chan *Task, 64),
		name:    name,
		role:    role,
		toolset: NewToolset(),
		buffer:  NewConversation(name),
		events:  NewEvents(),
		state:   WorkerStateCreating,
	}
}

/*
Start the worker and update the conversation buffer and state as needed.
*/
func (worker *Worker) Start() string {
	errnie.Trace()
	in := make(chan openai.ChatCompletionMessageParamUnion, 64)

	// Emit an event for worker start
	worker.events.channel <- Event{
		Timestamp: time.Now(),
		Type:      "WorkerStart",
		Message:   "Worker started processing.",
		WorkerID:  worker.name,
	}

	go func() {
		for {
			select {
			case <-worker.ctx.Done():
				worker.cancel()

				// Emit an event for worker cancellation
				worker.events.channel <- Event{
					Timestamp: time.Now(),
					Type:      "WorkerCancelled",
					Message:   "Worker context cancelled.",
					WorkerID:  worker.name,
				}

				return
			case task := <-worker.task:
				worker.state = WorkerStateWorking

				if task != nil {
					// Emit an event for task received
					worker.events.channel <- Event{
						Timestamp: time.Now(),
						Type:      "TaskReceived",
						Message:   task.sysStr + "\n\n" + task.usrStr,
						WorkerID:  worker.name,
					}

					// Use the task by resetting the conversation buffer
					worker.buffer.Reset(task)
				}

				// Initialize Executor with workerID and events channel
				executor := NewExecutor(worker.buffer, worker.toolset, worker.role, worker.name, in)
				executor.Start()
			case msg := <-in:
				worker.buffer.Update(msg)
			default:
				worker.events.channel <- Event{
					Timestamp: time.Now(),
					Type:      "WorkerState",
					Message:   fmt.Sprintf("Worker %s (%s) is %s.", worker.name, worker.role, worker.State()),
					WorkerID:  worker.name,
				}

				time.Sleep(3 * time.Second)
			}
		}
	}()

	return "Worker started"
}

func (worker *Worker) State() string {
	switch worker.state {
	case WorkerStateReady:
		return "ready"
	case WorkerStateWorking:
		return "working"
	case WorkerStateReviewing:
		return "reviewing"
	case WorkerStateDone:
		return "done"
	}

	return "unknown"
}
