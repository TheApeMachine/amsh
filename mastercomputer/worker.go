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

// Worker represents an agent that can perform tasks and communicate via the Queue.
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

// NewWorker provides a minimal, uninitialized Worker object.
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

// Initialize sets up the worker's context, queue registration, and starts the executor.
func (worker *Worker) Initialize() *Worker {
	errnie.Info("initializing %s %s (%s)", worker.buffer.Peek("role"), worker.buffer.Peek("id"), worker.name)

	// Generate a random float from 0.0 to 2.0, rounded to 1 decimal place
	temperature := utils.ToFixed(rand.Float64()*2.0, 1)

	// Store as a string inside the buffer.
	worker.buffer.Poke("temperature", fmt.Sprintf("%.1f", temperature))

	worker.state = WorkerStateInitializing
	worker.ctx, worker.cancel = context.WithCancel(worker.parentCtx)

	worker.memory = ai.NewMemory(worker.buffer.Peek("id"))
	worker.inbox, worker.err = worker.queue.Register(worker.name)

	if errnie.Error(worker.err) != nil {
		worker.state = WorkerStateZombie
		return nil
	}

	worker.queue.Subscribe(worker.name, worker.buffer.Peek("workload"))

	worker.manager.AddWorker(worker)
	worker.state = WorkerStateReady

	errnie.Note("worker %s initialized", worker.name)

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

				errnie.Note("%s <-[%s]- %s\n%s", worker.name, msg.Peek("role"), msg.Peek("scope"), msg.Peek("payload"))

				switch msg.Peek("role") {
				case "task":
					if worker.IsAllowed(WorkerStateAccepted) {
						msg.Poke("worker", worker.name)
						msg.Poke("temperature", worker.buffer.Peek("temperature"))

						NewExecutor(worker.ctx, msg)
					}
				case "unregister":
					worker.state = WorkerStateZombie
				}
			}
		}
	}()
}

func (worker *Worker) Stop() {
	worker.cancel()
}
