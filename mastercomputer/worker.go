package mastercomputer

import (
	"context"
	"fmt"
	"math/rand/v2"

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
	name      string
	manager   *twoface.WorkerManager
	err       error
}

// NewWorker provides a minimal, uninitialized Worker object.
func NewWorker(ctx context.Context, buffer *data.Artifact, manager *twoface.WorkerManager) *Worker {
	return &Worker{
		parentCtx: ctx,
		buffer:    buffer,
		state:     WorkerStateCreating,
		queue:     twoface.NewQueue(),
		name:      buffer.Peek("origin"),
		manager:   manager,
	}
}

// Initialize sets up the worker's context, queue registration, and starts the executor.
func (worker *Worker) Initialize() *Worker {
	errnie.Info("initializing %s %s (%s)", worker.buffer.Peek("role"), worker.ID(), worker.name)

	// Generate a random float from 0.0 to 2.0, rounded to 1 decimal place
	temperature := utils.ToFixed(rand.Float64()*2.0, 1)

	// Store as a string inside the buffer.
	worker.buffer.Poke("temperature", fmt.Sprintf("%.1f", temperature))

	worker.state = WorkerStateInitializing
	worker.ctx, worker.cancel = context.WithCancel(worker.parentCtx)

	worker.memory = ai.NewMemory(worker.ID())
	worker.inbox, worker.err = worker.queue.Register(worker.ID())

	if errnie.Error(worker.err) != nil {
		worker.state = WorkerStateZombie
		return nil
	}

	worker.queue.Subscribe(worker.ID(), worker.buffer.Peek("workload"))

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
		return worker.queue.Unregister(worker.ID())
	}

	return nil
}

func (worker *Worker) Ctx() context.Context {
	return worker.ctx
}

func (worker *Worker) Manager() *twoface.WorkerManager {
	return worker.manager
}

func (worker *Worker) ID() string {
	return worker.buffer.Peek("id")
}

func (worker *Worker) Name() string {
	return worker.name
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
						worker.ID(), "state", "notok", []byte(worker.buffer.Peek("payload")),
					))

					errnie.Warn("worker %s inbox channel closed", worker.ID())
					continue
				}

				errnie.Info("%s <-[%s]- %s", worker.name, msg.Peek("role"), msg.Peek("scope"))
				errnie.Note("[PAYLOAD]\n%s\n[/PAYLOAD]", msg.Peek("payload"))

				NewMessaging(worker).Reply(msg)

				errnie.Info("worker %s state: %s", worker.name, worker.state)

				if worker.state == WorkerStateAccepted {
					worker.state = WorkerStateBusy
					exec := NewExecutor(worker.ctx, worker)
					exec.Initialize()
					exec.Execute(msg)
					exec.Verify()
					worker.state = WorkerStateDone
				}
			}
		}
	}()
}

func (worker *Worker) Stop() {
	worker.cancel()
}

func (worker *Worker) Call(args map[string]any, owner twoface.Process) (string, error) {
	builder := NewBuilder(owner.Ctx(), owner.Manager())
	reasoner := builder.NewWorker(builder.getRole(args["toolset"].(string)))
	reasoner.buffer.Poke("system", args["system"].(string))
	reasoner.buffer.Poke("user", args["user"].(string))
	reasoner.buffer.Poke("workload", args["toolset"].(string))
	reasoner.buffer.Poke("parent", owner.ID())
	reasoner.Start()
	return utils.ReplaceWith(`
	[WORKER {name}]
	  STATE   : {state}
	  WORKLOAD: {workload}
	  PARENT  : {parent}
	[/WORKER]
	`, [][]string{
		{"name", worker.name},
		{"state", worker.state.String()},
		{"workload", args["toolset"].(string)},
		{"parent", owner.ID()},
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
