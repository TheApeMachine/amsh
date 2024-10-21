package mastercomputer

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

/*
Worker represents a blank agent that can adopt any role or workload that is assigned to it.
By setting the system prompt, user prompt, and assigning a toolset, the worker is flexible enough
to be the only agentic type. Finally, given that a worker is also a tool, workers that are assigned
the worker tool can make their own workers and delegate work.
*/
type Worker struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	buffer    *data.Artifact
	memory    *ai.Memory
	state     WorkerState
	process   *Process
	pool      *twoface.Pool
	queue     *twoface.Queue
	inbox     chan data.Artifact
	name      string
	manager   *Manager
	err       error
}

/*
NewWorker provides a minimal, uninitialized worker object. The buffer artifact is used to
initialize the worker's configuration, and an essential part of data transfer.
*/
func NewWorker(ctx context.Context, buffer *data.Artifact, manager *Manager) *Worker {
	return &Worker{
		parentCtx: ctx,
		buffer:    buffer,
		state:     WorkerStateCreating,
		pool:      twoface.NewPool(),
		queue:     twoface.NewQueue(),
		name:      buffer.Peek("origin"),
		manager:   manager,
	}
}

/*
Initialize sets up the worker's context, queue registration, and initializes the memory.
This is the preparation phase for the worker to be ready to receive and process messages.
*/
func (worker *Worker) Initialize() *Worker {
	worker.state = WorkerStateInitializing
	worker.ctx, worker.cancel = context.WithCancel(worker.parentCtx)

	temperature := utils.ToFixed(rand.Float64()*1.0, 1)
	worker.buffer.Poke("temperature", fmt.Sprintf("%.1f", temperature))

	worker.memory = ai.NewMemory(worker.name)
	worker.inbox, worker.err = worker.queue.Register(worker.name)

	if errnie.Error(worker.err) != nil {
		worker.state = WorkerStateZombie
		return nil
	}

	worker.queue.Subscribe(worker.name, worker.buffer.Peek("workload"))
	worker.manager.AddWorker(worker)
	worker.state = WorkerStateReady

	return worker
}

// Close shuts down the worker and cleans up resources.
func (worker *Worker) Close() error {
	if worker.cancel != nil {
		worker.cancel()
	}

	if worker.queue != nil {
		return worker.queue.Unregister(worker.name)
	}

	return nil
}

// Start the worker and listen for messages from the queue.
func (worker *Worker) Start() {
	go func() {
		for {
			select {
			case <-worker.parentCtx.Done():
				return
			case <-worker.ctx.Done():
				return
			case msg, ok := <-worker.inbox:
				if !ok {
					worker.state = WorkerStateError
					worker.queue.PubCh <- *data.New(
						worker.name, "state", "error", []byte(worker.buffer.Peek("payload")),
					)
					continue
				}

				var proceed bool

				errnie.Note("%s (%s) <-[%s]- %s", worker.name, worker.buffer.Peek("role"), msg.Peek("role"), msg.Peek("origin"))

				// Handle the message, this will set the worker's process flow,
				// or ignore it when the worker is not ready.
				if proceed = NewMessaging(worker, &msg).Process(); proceed {
					var step map[string]string

					// Pop the first step from the process flow.
					step, worker.process.flow = worker.process.flow[0], worker.process.flow[1:]

					if len(worker.process.flow) == 0 {
						worker.process = nil
					}

					// Set the worker state according to the process flow step.
					worker.state = worker.StateByKey(step["state"])

					// Set the system prompt for the message.
					msg.Poke("system", worker.buffer.Peek("system"))

					// Set the user prompt for the message.
					msg.Poke("user", strings.Join([]string{worker.buffer.Peek("user"), step["user"]}, "\n\n"))

					iterations, err := strconv.Atoi(step["iterations"])
					if errnie.Error(err) != nil {
						iterations = 1
					}

					// Execute the message, this will update the message's payload.
					if proceed = NewExecutor(worker, &msg).Do(iterations); proceed {
						// Update the message according to the process flow step.
						NewMessaging(worker, &msg).Update(step)

						// Publish the message back onto the queue.
						errnie.Note("%s (%s) -[%s]-> %s", worker.name, worker.buffer.Peek("role"), msg.Peek("role"), msg.Peek("scope"))
						worker.queue.PubCh <- msg
					}
				}

				// Reset the worker state to ready, so it can process the next message.
				worker.state = WorkerStateReady
			}
		}
	}()
}

func (worker *Worker) Stop() {
	worker.cancel()
}
