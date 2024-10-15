package mastercomputer

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
)

/*
Worker is an agent type that can become many things, depending on how it is prompted.
By setting the system and user prompts, as well as providing a toolset, we can
instruct the worker to perform many different tasks.
A worker is itself also a tool, and can create other workers, and thus form a
tree of workers.

It carries an Artifact type as a buffer, which will hold the current context of
the worker, which always includes the syatem an user prompt, so it never loses scope
but the assistant messages need to be managed to not overflow the context window.
*/
type Worker struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	wg        *sync.WaitGroup
	err       error
	buffer    data.Artifact
	memory    *ai.Memory
	State     WorkerState
	queue     *twoface.Queue
	inbox     chan data.Artifact
	OK        bool
}

/*
NewWorker provides a minimal, uninitialized Worker object. We pass in a
context for cancellation purposes, and an Artifact so we can transfer
data over if we need to.
*/
func NewWorker(ctx context.Context, wg *sync.WaitGroup, buffer data.Artifact) *Worker {
	errnie.Trace()

	return &Worker{
		parentCtx: ctx,
		wg:        wg,
		buffer:    buffer,
		State:     WorkerStateCreating,
		OK:        false,
	}
}

/*
Initialize the worker, setting up the context, ID, memory, and queue.
It is essential that this is called before the worker can be used.
*/
func (worker *Worker) Initialize() *Worker {
	errnie.Info("initializing worker %s", worker.buffer.Peek("id"))

	worker.State = WorkerStateInitializing
	worker.ctx, worker.cancel = context.WithCancel(worker.parentCtx)

	// Registering the worker with the queue allows it to send and receive
	// messages, enabling worker communication. The queue itself is an
	// ambient context, so all workers speak to the same queue instance.
	worker.queue = twoface.NewQueue()

	// Inbox is the incoming message channel for the worker, and without it
	// the worker is essentially rogue, and not functional.
	if worker.inbox, worker.err = worker.queue.Register(
		worker.buffer.Peek("id"),
	); errnie.Error(worker.err) != nil {
		worker.State = WorkerStateZombie
	}

	if worker.State == WorkerStateZombie {
		errnie.Error(errors.New("[" + worker.buffer.Peek("id") + "] went zombie"))
		// Kill it with fire!
		worker.Close()
	}

	worker.State = WorkerStateReady
	worker.OK = true

	executor := NewExecutor(worker.ctx, worker)
	executor.Initialize()

	_, err := io.Copy(worker, executor)

	if errnie.Error(err) != nil {
		worker.State = WorkerStateZombie
		worker.OK = false
	}

	return worker
}

/*
Error implements the error interface for the worker.
*/
func (worker *Worker) Error() string {
	errnie.Trace()
	return worker.err.Error()
}

/*
Read implements the io.Reader interface for the worker.
In this case, it is reading from the buffer payload, which
is a slightly different use-case from the behavior that
Artifacts naturally implement, also being io.ReadWriteClosers.
*/
func (worker *Worker) Read(p []byte) (n int, err error) {
	errnie.Trace()

	if !worker.OK || worker.State != WorkerStateReady {
		return 0, io.ErrNoProgress
	}

	n = copy(p, worker.buffer.Peek("payload"))
	return n, io.EOF
}

/*
Write implements the io.Writer interface for the worker.
In this case, it is writing to the buffer payload, which
is a slightly different use-case from the behavior that
Artifacts naturally implement, also being io.ReadWriteClosers.
*/
func (worker *Worker) Write(p []byte) (n int, err error) {
	errnie.Trace()

	if _, err = worker.memory.Read(p); err != nil && err != io.EOF {
		return 0, err
	}

	if !worker.OK || worker.State != WorkerStateReady {
		return 0, io.ErrNoProgress
	}

	return len(p), worker.buffer.SetPayload(p)
}

/*
Close the worker down and make sure it cleans up all its resources.
*/
func (worker *Worker) Close() error {
	errnie.Trace()
	worker.cancel()
	return nil
}
