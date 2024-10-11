package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

type WorkerState uint

const (
	WorkerStateCreating WorkerState = iota
	WorkerStateInitializing
	WorkerStateReady
	WorkerStateRunning
	WorkerStateFinished
)

type Worker struct {
	parentCtx context.Context
	ctx       context.Context
	err       error
	buffer    data.Artifact
	memory    *ai.Memory
	state     WorkerState
	queue     *twoface.Queue
	inbox     chan data.Artifact
	ID        string
	Function  WorkerTool
}

func NewWorker(ctx context.Context, buffer data.Artifact) *Worker {
	errnie.Trace()

	return &Worker{
		ctx:    ctx,
		buffer: buffer,
		state:  WorkerStateCreating,
	}
}

func (worker *Worker) Initialize(ctx context.Context) {
	worker.state = WorkerStateInitializing
	worker.queue = twoface.NewQueue()
	worker.queue.Register(worker.ID)
	worker.inbox, _ = worker.queue.Subscribe(worker.ID, worker.ID)

	listener := twoface.NewListener(worker.ctx, worker.inbox)
	listener.Messages(worker.inboxCallback)
}

func (worker *Worker) Read(p []byte) (n int, err error) {
	return worker.memory.Read(p)
}

func (worker *Worker) Write(p []byte) (n int, err error) {
	return worker.memory.Write(p)
}

func (worker *Worker) Close() error {
	return nil
}

func (worker *Worker) SendMessage(topic string, message data.Artifact) {
	worker.queue.Publish(topic, message)
}

func (worker *Worker) inboxCallback(msg data.Artifact) {
	io.Copy(worker.memory, msg)
}

// Implement the Job interface
func (worker *Worker) Process(ctx context.Context) error {
	errnie.Trace()
	worker.state = WorkerStateRunning

	params := GetParams(
		worker.buffer.Peek("system"),
		worker.buffer.Peek("user"),
		[]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.String("worker"),
					Description: openai.String("Create any type of worker, by providing a system prompt, a user prompt, and a toolset"),
					Parameters:  openai.F(WorkerToolSchema()),
				}),
			},
		},
	)

	for {
		params = worker.handleToolCalls(worker.printResponse(
			NewCompletion(worker.ctx).Execute(worker.ctx, params),
		), params)
	}

	return nil
}

func (worker *Worker) printResponse(response openai.ChatCompletion) openai.ChatCompletionMessage {
	errnie.Trace()

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
	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		return params
	}

	var (
		args map[string]interface{}
		out  string
	)

	params.Messages.Value = append(params.Messages.Value, message)
	for _, toolCall := range message.ToolCalls {
		errnie.Raw(toolCall)

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
				wrkr.state,
			)

			utils.PrettyJSON(args)
			errnie.Debug("ToolCall: %s %s", toolCall.ID, out)
			params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, out))
		case "environment":
			out = "not implemented"
		default:
			out = fmt.Sprintf("unknown tool call: %s", toolCall.Function.Name)
			errnie.Warn(out)
		}

		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); errnie.Error(err) != nil {
			errnie.Error(err)
		}
	}

	return params
}
