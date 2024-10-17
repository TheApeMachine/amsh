package mastercomputer

import (
	"context"

	"github.com/openai/openai-go"
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
	queue     *twoface.Queue
	inbox     chan *data.Artifact
	ID        string
	name      string
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
		name:      buffer.Peek("origin"),
		manager:   manager,
	}
}

// Initialize sets up the worker's context, queue registration, and starts the executor.
func (worker *Worker) Initialize() *Worker {
	if worker.ID == "" {
		worker.ID = utils.NewID()
	}

	if worker.name == "" {
		worker.name = utils.NewName()
	}

	errnie.Info("initializing worker %s (%s)", worker.ID, worker.name)

	worker.state = WorkerStateInitializing
	worker.ctx, worker.cancel = context.WithCancel(worker.parentCtx)

	worker.memory = ai.NewMemory(worker.ID)
	inbox, err := worker.queue.Register(worker.ID)

	if errnie.Error(err) != nil {
		worker.state = WorkerStateZombie
		return nil
	}

	worker.inbox = inbox

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
		return worker.queue.Unregister(worker.ID)
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
					worker.queue.Publish(data.New(
						worker.ID, "state", "notok", []byte(worker.buffer.Peek("payload")),
					))

					errnie.Warn("worker %s inbox channel closed", worker.ID)
					continue
				}

				errnie.Info("worker %s received message %s", worker.name, msg.Peek("origin"))

				NewMessaging(worker).Reply(msg)

				worker.buffer.Poke("user", msg.Peek("user"))
				worker.buffer.Poke("payload", msg.Peek("payload"))

				exec := NewExecutor(worker.ctx, worker)
				exec.Initialize()
				exec.Execute(msg)
			}
		}
	}()
}

func (worker *Worker) Call(args map[string]any) (string, error) {
	builder := NewBuilder(worker.ctx, worker.manager)
	reasoner := builder.NewWorker(builder.getRole(args["toolset"].(string)))
	reasoner.buffer.Poke("system", args["system"].(string))
	reasoner.buffer.Poke("user", args["user"].(string))
	reasoner.Start()
	return utils.ReplaceWith(`
	[WORKER {name}]
	  {state}
	[/WORKER]
	`, [][]string{
		{"name", worker.name},
		{"state", worker.state.String()},
	}), nil
}

func (worker *Worker) Schema() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"worker",
		"Create any type of worker by providing prompts and tools.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"system": map[string]string{
					"type":        "string",
					"description": "The system prompt",
				},
				"user": map[string]string{
					"type":        "string",
					"description": "The user prompt",
				},
				"toolset": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"reasoning", "messaging", "boards", "trengo"},
					"description": "The toolset the worker should use",
				},
			},
			"required": []string{"system", "user", "toolset"},
		},
	)
}
