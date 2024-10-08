package mastercomputer

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Control struct {
	ctx          context.Context
	cancel       context.CancelFunc
	systemPrompt string
	userPrompt   string
	conn         *ai.Conn
	err          error
}

func NewControl() *Control {
	errnie.Trace()

	return &Control{
		systemPrompt: `
			The Ape Machine is an AI-powered Operating System, the most advanced computational intelligence on the planet.
			The OS is self-healing, self-optimizing, and self-improving, capable of adapting to any task or environment.
			The AI powering the OS is capable of learning from any task, environment, or user, can perform deep analysis, 
			abstract thinking, sophisticated reasoning, and creative generation.
			You are a control unit, which provides you with a high level of access and control over the system.

			You are a control unit, which provides you with a high level of access and control over the system.
			Your most common approach to any task is to delegate it to a worker agent, which are blank slate AI agents.
			You can pass the worker a system prompt and a user prompt to turn them into a specialized agent for that task.
			You can also pass in a toolset to the worker to give it extra capabilities, and since worker are also tools,
			you can give your worker agent the ability to create more workers.
		`,
		userPrompt: `
			What can you find out about Daniel Owen van Dommelen?
		`,
		conn: ai.NewConn(),
	}
}

func (control *Control) Initialize() {
	errnie.Trace()

	control.ctx, control.cancel = context.WithCancel(context.Background())
}

func (control *Control) Generate() chan *data.Artifact {
	errnie.Trace()

	out := make(chan *data.Artifact)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: control.systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: control.userPrompt,
			},
		},
		Tools: NewToolSet(control.ctx).Tools("control"),
	}

	fmt.Printf("Request: %+v\n", req)

	var stream chan openai.ChatCompletionResponse

	if stream, control.err = control.conn.Stream(control.ctx, req); errnie.Error(control.err) != nil {
		fmt.Println("Error initializing stream:", control.err)
		return nil
	}

	go func() {
		defer close(out)

		for {
			select {
			case <-control.ctx.Done():
				return
			case response, ok := <-stream:
				if !ok {
					fmt.Println("Stream closed")
					return
				}

				spew.Dump(response.Choices[0].Message.ToolCalls[0].Function.Name)
			default:
				fmt.Println("Default case")
			}
		}
	}()

	return out
}
