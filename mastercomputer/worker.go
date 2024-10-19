package mastercomputer

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

/*
Worker represents a blank agent that can adopt any role or workload that is assigned to it.
By setting the system prompt, user prompt, and assigning a toolset, the worker is flexible enough
to be the only agentic type. Finally, given that a worker is also a tool, worker who are assigned
the worker tool, can make their own workers and delegate work.
*/
type Worker struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	buffer    *data.Artifact
	memory    *ai.Memory
	state     WorkerState
	pool      *twoface.Pool
	queue     *twoface.Queue
	inbox     chan *data.Artifact
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

	worker.memory = ai.NewMemory(worker.buffer.Peek("id"))
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
					worker.state = WorkerStateNotOK
					worker.queue.PubCh <- data.New(
						worker.name, "state", "notok", []byte(worker.buffer.Peek("payload")),
					)

					continue
				}

				errnie.Debug("%s <-[%s]- %s\n%s", worker.name, msg.Peek("role"), msg.Peek("scope"), msg.Peek("payload"))

				if msg := NewMessaging(worker, msg).Process(); msg != nil {
					worker.queue.PubCh <- NewExecutor(worker, msg).Do()
				}

				worker.NewState(worker.StateByKey("ready"))
			}
		}
	}()
}

func (worker *Worker) Stop() {
	worker.cancel()
}
