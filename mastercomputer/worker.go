package mastercomputer

import (
	"context"
	"log"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
)

// Worker represents an agent that can perform tasks and communicate via the Queue.
type Worker struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	buffer    *data.Artifact
	memory    *ai.Memory
	state     WorkerState
	queue     *twoface.Queue
	inbox     chan *data.Artifact
	ID        string
	manager   *WorkerManager
}

// NewWorker provides a minimal, uninitialized Worker object.
func NewWorker(ctx context.Context, buffer *data.Artifact, manager *WorkerManager) *Worker {
	return &Worker{
		parentCtx: ctx,
		buffer:    buffer,
		state:     WorkerStateCreating,
		queue:     twoface.NewQueue(),
		ID:        buffer.Peek("id"),
		manager:   manager,
	}
}

// Initialize sets up the worker's context, queue registration, and starts the executor.
func (w *Worker) Initialize() error {
	log.Printf("Initializing worker %s", w.ID)
	w.state = WorkerStateInitializing
	w.ctx, w.cancel = context.WithCancel(w.parentCtx)

	w.memory = ai.NewMemory(w.ID)
	defer w.memory.Close()

	// Register the worker with the queue.
	inbox, err := w.queue.Register(w.ID)
	if err != nil {
		w.state = WorkerStateZombie
		return err
	}
	w.inbox = inbox

	// Add the worker to the manager
	w.manager.AddWorker(w)

	w.state = WorkerStateReady
	w.listenForMessages()

	// Start the executor
	executor := NewExecutor(w.ctx, w)
	if err := executor.Initialize(); err != nil {
		return err
	}
	go executor.Run()

	log.Printf("Worker %s initialized", w.ID)

	// Wait until the context is done
	<-w.ctx.Done()

	// Indicate that the worker is done
	w.manager.RemoveWorker(w.ID)
	return nil
}

// Close shuts down the worker and cleans up resources.
func (w *Worker) Close() error {
	if w.cancel != nil {
		w.cancel()
	}
	if w.queue != nil {
		return w.queue.Unregister(w.ID)
	}
	return nil
}

// listenForMessages listens to the inbox and processes messages.
func (w *Worker) listenForMessages() {
	go func() {
		for {
			select {
			case <-w.ctx.Done():
				return
			case msg, ok := <-w.inbox:
				if !ok {
					return
				}
				w.handleMessage(msg)
			}
		}
	}()
}

// handleMessage processes incoming messages.
func (w *Worker) handleMessage(msg *data.Artifact) {
	log.Printf("Worker %s received message from %s", w.ID, msg.Peek("origin"))
	// Implement message handling logic based on the message type and content.
	// For example, pick up broadcast messages if in the Ready state.
	if w.state == WorkerStateReady && msg.Peek("scope") == "broadcast" {
		w.NewState(WorkerStateBusy)
		// Acknowledge picking up the workload
		response := data.New(w.ID, "message", "acknowledgment", []byte("Workload accepted"))
		response.Poke("origin", w.ID)
		response.Poke("scope", msg.Peek("origin"))
		w.queue.Publish(response)
		// Process the workload
		// ...
		w.NewState(WorkerStateDone)
	}
}
