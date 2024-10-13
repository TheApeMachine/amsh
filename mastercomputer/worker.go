package mastercomputer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/container"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

type WorkerState uint

const (
	// WorkerStateCreating is the starting state of the worker, indicating that it has not been initialized yet.
	WorkerStateCreating WorkerState = iota

	// WorkerStateInitializing indicates we are intializing, and nothing has gone wrong yet.
	WorkerStateInitializing

	// WorkerStateReady indicates that the worker is ready to take on work.
	WorkerStateReady

	// WorkerStateAcknowledged indicates the worker received a message and is sending ACK to the sender.
	WorkerStateAcknowledged

	// WorkerStateAccepted indicates the worker aceepted a workload and is now the owner of it.
	WorkerStateAccepted

	// WorkerStateRejected indicates the worker rejected a workload.
	WorkerStateRejected

	// WorkerStateBusy indicates the worker is currently actively performing work.
	WorkerStateBusy

	// WorkerStateWaiting indicates the worker is actively performing work, or about to, but waiting for additional input.
	WorkerStateWaiting

	// WorkerStateDone indicates the worker has completed the work it was previously busy with.
	WorkerStateDone

	// WorkerStateError indicates the worker has experienced an error.
	WorkerStateError

	// WorkerStateFinished indicates the worker has been deallocated and is shutting down.
	WorkerStateFinished

	// WorkerStateZombie indicates the worker has experienced a fatal error and is shutting down.
	WorkerStateZombie
)

type Worker struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	err       error
	buffer    data.Artifact
	memory    *ai.Memory
	State     WorkerState
	queue     *twoface.Queue
	inbox     chan data.Artifact
	ID        string
	OK        bool
}

/*
NewWorker provides a minimal, uninitialized Worker object. We pass in a
context for cancellation purposes, and an Artifact so we can transfer
data over if we need to.
*/
func NewWorker(ctx context.Context, buffer data.Artifact) *Worker {
	errnie.Trace()

	return &Worker{
		parentCtx: ctx,
		buffer:    buffer,
		State:     WorkerStateCreating,
		OK:        false,
	}
}

func (worker *Worker) Initialize() *Worker {
	errnie.Trace()

	worker.ctx, worker.cancel = context.WithCancel(worker.parentCtx)
	worker.ID = utils.NewID()

	worker.State = WorkerStateInitializing
	worker.queue = twoface.NewQueue()
	worker.queue.Register(worker.ID)

	if worker.inbox, worker.err = worker.queue.Register(worker.ID); errnie.Error(worker.err) != nil {
		worker.State = WorkerStateZombie
	}

	if worker.State == WorkerStateZombie {
		errnie.Error(errors.New("[" + worker.ID + "] went zombie"))
	}

	worker.State = WorkerStateReady
	worker.OK = true
	errnie.Info("worker: %s OK and ready", worker.ID)

	return worker
}

/*
Test sends a test message to the worker, and returns the response.
*/
func (worker *Worker) Test(msg data.Artifact) string {
	errnie.Trace()
	msg.Poke("origin", worker.ID)
	worker.queue.Publish(msg)
	return ""
}

/*
This method will block until an error is available, and then return it.
Any outside method that creates a worker can immediately call this method
meaning, they can often just put that in their return, and it will block
unless there is an error.
*/
func (worker *Worker) Error() string {
	errnie.Trace()
	return worker.err.Error()
}

func (worker *Worker) Read(p []byte) (n int, err error) {
	errnie.Trace()
	return worker.buffer.Read(p)
}

func (worker *Worker) Write(p []byte) (n int, err error) {
	errnie.Trace()

	if !worker.OK || worker.State != WorkerStateReady {
		return 0, io.ErrNoProgress
	}

	return worker.buffer.Write(p)
}

func (worker *Worker) Close() error {
	errnie.Trace()
	worker.cancel()
	return nil
}

// Implement the Job interface
func (worker *Worker) Process() error {
	errnie.Trace()

	if worker.State != WorkerStateReady {
		return errors.New("worker not ready")
	}

	go func() {
		for {
			select {
			case <-worker.parentCtx.Done():
				worker.State = WorkerStateFinished
				return
			case <-worker.ctx.Done():
				worker.State = WorkerStateFinished
				return
			case msg := <-worker.inbox:
				errnie.Info("worker: %s received message", worker.ID)
				utils.PrettyJSON(msg)

				var (
					final  string
					reason string
				)

				worker.State, reason = worker.ReplyState(msg)

				worker.queue.Publish(data.New(
					worker.ID, "response", msg.Peek("origin"), []byte(utils.MessageTemplate(
						"RE: "+msg.Peek("id"),
						"reply",
						msg.Peek("topic"),
						final+"; "+reason,
					)),
				))

				switch worker.State {
				case WorkerStateAccepted:
					worker.Work(msg)
				}
			default:
				time.Sleep(1000 * time.Millisecond)
				worker.State = WorkerStateReady
				errnie.Info("worker: %s is OK and ready", worker.ID)
			}
		}
	}()

	return nil
}

/*
MsgState looks at the message and determines the response to send back to the
queue.
*/
func (worker *Worker) ReplyState(msg data.Artifact) (WorkerState, string) {
	errnie.Trace()

	message, err := NewCompletion(worker.ctx).Execute(worker.ctx, GetParams(
		worker.buffer.Peek("system"),
		worker.buffer.Peek("user")+"\n\n"+msg.Peek("payload"),
		openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        openai.F("messaging"),
			Description: openai.F("Available messaging formats"),
			Schema:      openai.F(GenerateSchema[format.Messaging]()),
			Strict:      openai.Bool(true),
		},
		NewToolset("none").tools,
	))

	if errnie.Error(err) != nil {
		return WorkerStateError, err.Error()
	}

	response, err := utils.JSONtoMap(message.Choices[0].Message.Content)
	if errnie.Error(err) != nil {
		return WorkerStateError, err.Error()
	}

	if final, ok := response["final_response"].(format.Reply); ok {
		if final.Accepted {
			return WorkerStateAccepted, final.Reason
		}

		return WorkerStateRejected, final.Reason
	}

	return WorkerStateRejected, "no response"
}

var formatMap = map[string]openai.ResponseFormatJSONSchemaJSONSchemaParam{
	"reasoning": {
		Name:        openai.F("reasoning"),
		Description: openai.F("Available reasoning strategies"),
		Schema:      openai.F(GenerateSchema[format.Strategy]()),
		Strict:      openai.Bool(false),
	},
	"environment": {
		Name:        openai.F("environment"),
		Description: openai.F("Available environment commands"),
		Schema:      openai.F(GenerateSchema[format.Environment]()),
		Strict:      openai.Bool(true),
	},
}

/*
Work is an infinite self-prompting loop that will continue until the worker decides
it is done, or another breaking condition is met.
*/
func (worker *Worker) Work(msg data.Artifact) {
	errnie.Trace()
	worker.State = WorkerStateBusy

	system, user := utils.ComposedMessage(
		worker.ID, worker.buffer,
	)

	params := GetParams(
		system, user,
		formatMap[worker.buffer.Peek("format")],
		NewToolset(worker.buffer.Peek("toolset")).tools,
	)

	for {
		if worker.State == WorkerStateFinished {
			// The worker has been killed, send a cancellation to our own context so we can exit cleanly.
			// This will also cascade to any object that has our context as its parent.
			worker.Close()
			return
		}

		if worker.State == WorkerStateWaiting {
			// We are waiting for additional input, so we wait and try again.
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		if worker.State == WorkerStateDone {
			// The objective has been completed, so we exit the prompting loop
			break
		}

		params = worker.handleToolCalls(worker.printResponse(
			NewCompletion(worker.ctx).Execute(worker.ctx, params),
		), params)
	}
}

func (worker *Worker) printResponse(response *openai.ChatCompletion, err error) openai.ChatCompletionMessage {
	errnie.Trace()

	if errnie.Error(err) != nil {
		worker.State = WorkerStateFinished
		return openai.ChatCompletionMessage{}
	}

	reasoning := map[string]any{}

	if worker.err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &reasoning); worker.err != nil {
		errnie.Error(worker.err)
		return response.Choices[0].Message
	}

	utils.PrettyJSON(reasoning)
	return response.Choices[0].Message
}

func (worker *Worker) handleToolCalls(
	message openai.ChatCompletionMessage, params openai.ChatCompletionNewParams,
) openai.ChatCompletionNewParams {
	errnie.Trace()

	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		return params
	}

	var (
		args map[string]interface{}
		out  string
	)

	params.Messages.Value = append(params.Messages.Value, message)
	for _, toolCall := range message.ToolCalls {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); errnie.Error(err) != nil {
			out = "error unmarshalling arguments"
		}

		switch toolCall.Function.Name {
		case "worker":
			wrkr := NewWorker(
				worker.ctx,
				data.New(
					worker.ID, "prompt", "task", nil,
				).Poke(
					"system", args["system"].(string),
				).Poke(
					"user", args["user"].(string),
				).Poke(
					"toolset", args["toolset"].(string),
				),
			)

			out = fmt.Sprintf(
				"[%s @ %s]\nSYSTEM: %s\nUSER: %s\nSTATUS: %d\n",
				wrkr.ID,
				time.Now().Format(time.RFC3339),
				wrkr.buffer.Peek("system"),
				wrkr.buffer.Peek("user"),
				wrkr.State,
			)

			utils.PrettyJSON(args)
			errnie.Debug("ToolCall: %s %s", toolCall.ID, out)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		case "environment":
			worker.handleEnvironment(args)
		default:
			utils.PrettyJSON(args)
			errnie.Debug("ToolCall: %s %s", toolCall.ID, out)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		}

		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); errnie.Error(err) != nil {
			errnie.Error(err)
		}
	}

	return params
}

/*
handleEnvironment sets up an environment for the worker to use, then places the
worker in a separate prompting loop designed to interact with the environment.
*/
func (worker *Worker) handleEnvironment(args map[string]interface{}) {
	errnie.Trace()

	// Ensure the container is built and running
	builder, err := container.NewBuilder()
	if errnie.Error(err) != nil {
		worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError setting up environment: "+err.Error())
		return
	}

	if err = builder.BuildImage(worker.ctx, ".", worker.ID+"-environment"); errnie.Error(err) != nil {
		worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError building environment image: "+err.Error())
		return
	}

	runner, err := container.NewRunner()
	if errnie.Error(err) != nil {
		worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError creating container runner: "+err.Error())
		return
	}

	in, out, err := runner.RunContainer(
		worker.ctx,
		worker.ID+"-environment",
		[]string{"/bin/bash"},
		worker.ID, viper.GetViper().GetString("tools.environment.instructions"),
	)

	if errnie.Error(err) != nil {
		worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError running container: "+err.Error())
		return
	}

	env := NewEnvironment(worker.ctx, in, out)
	defer env.Close()

	// Now we can put the worker in a loop to interact with the environment.
	for {
		response, err := NewCompletion(worker.ctx).Execute(worker.ctx, GetParams(
			worker.buffer.Peek("system"),
			worker.buffer.Peek("user"),
			formatMap["environment"],
			NewToolset("none").tools,
		))

		if errnie.Error(err) != nil {
			worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError getting AI response: "+err.Error())
			continue
		}

		// Process the AI's response
		var envCommand format.Environment
		if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &envCommand); errnie.Error(err) != nil {
			worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError parsing AI response: "+err.Error())
			continue
		}

		// Execute the command in the environment
		if _, err := env.Write([]byte(envCommand.Command + "\n")); errnie.Error(err) != nil {
			worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError writing command to environment: "+err.Error())
			continue
		}

		// Read the output
		output := make([]byte, 4096)
		n, err := env.Read(output)
		if err != nil && err != io.EOF {
			errnie.Error(err)
			worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nError reading from environment: "+err.Error())
			continue
		}

		// Update the user prompt with the command output
		worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nCommand: "+envCommand.Command+"\nOutput: "+string(output[:n]))

		// Check if we're done
		if envCommand.Command == "exit" {
			break
		}
	}

	// Final message to indicate the environment interaction is complete
	worker.buffer.Poke("user", worker.buffer.Peek("user")+"\n\nEnvironment interaction complete.")
}
