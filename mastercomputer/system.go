package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

type System struct {
	ctx              context.Context
	cancel           context.CancelFunc
	err              error
	queue            *twoface.Queue
	messages         []data.Artifact
	Prompt           string
	Persona          string
	Responsibilities string
	Memory           *ai.Memory
	BootSequence     []string
	Toolset          *Toolset
	ID               string
	I                chan string
	O                chan string
}

func NewSystem() *System {
	ID := utils.NewID()

	return &System{
		ID:       ID,
		queue:    twoface.NewQueue(),
		messages: []data.Artifact{},
		Prompt: `
		The Ape Machine is an AI-powered Operating System, designed to handle any task, environment, or user.

		PRIME OBJECTIVE: To reason beyond reason, think deep thoughts, be generous with your time, and be anything the user needs.
		`,
		Persona: fmt.Sprintf(`You are %s, and you are a System, making you the context other sub-systems, components, and processes run within.`, ID),
		Responsibilities: `
		Manage your context efficiently, striving for optimal performance, and minimal catastrophic failure, and keep the Prime Objective in mind, always.

		Examine your tools, they are everything you need.
		`,
		I: make(chan string),
		O: make(chan string),
	}
}

func (system *System) Initialize() error {
	errnie.Trace()

	system.ctx, system.cancel = context.WithCancel(context.Background())
	system.Memory = ai.NewMemory()

	return nil
}

func (system *System) Generate() {
	errnie.Trace()

	broadcast := system.queue.Register(system.ID)

	go func() {
		for {
			select {
			case <-system.ctx.Done():
				return
			case message := <-broadcast:
				system.messages = append(system.messages, message)
			case prompt := <-system.I:
				errnie.Info("%s received prompt: %s", system.ID, prompt)

				// Add messages to the prompt context.
				prompt += "\n\n[MESSAGES]\n"
				for _, message := range system.messages {
					if origin, err := message.Origin(); err == nil {
						if role, err := message.Role(); err == nil {
							if scope, err := message.Scope(); err == nil {
								prompt += fmt.Sprintf("[%s @ %s]: %s, %s\n", origin, time.Now().Format("15:04:05"), role, scope)
							}

							if payload, err := message.Payload(); err == nil {
								prompt += fmt.Sprintf("%s\n", string(payload))
							}
						}
					}
				}
				prompt += "\n[/MESSAGES]\n\n"

				response := NewCompletion(system.ctx).Execute(
					strings.Join([]string{system.Prompt, system.Persona, system.Responsibilities}, "\n\n"),
					prompt,
					"system",
					format.NewChainOfThought(),
				)

				message := response.Choices[0].Message

				if len(message.ToolCalls) > 0 {
					for _, toolCall := range message.ToolCalls {
						system.Memory.ShortTerm = append(system.Memory.ShortTerm, fmt.Sprintf("TOOL CALL: %s, ARGUMENTS: %s", toolCall.Function.Name, toolCall.Function.Arguments))
						system.useTool(toolCall)
					}
				}

				if message.Content != "" {
					system.Memory.ShortTerm = append(system.Memory.ShortTerm, message.Content)

					var chainOfThought format.ChainOfThought

					if errnie.Error(json.Unmarshal([]byte(message.Content), &chainOfThought.Template)) != nil {
						return
					}

					utils.BeautifyChainOfThought(system.ID, chainOfThought)

					if strings.ToUpper(chainOfThought.Template.Action) == "TERMINATE" {
						return
					}
				}
			default:
				var nextMsg data.Artifact
				if len(system.messages) > 0 {
					nextMsg, system.messages = system.messages[0], system.messages[1:]
					system.Memory.ShortTerm = append(system.Memory.ShortTerm, nextMsg.String())
				}

				system.I <- "Check the context and decide what to do next, or answer NOOP to do nothing, or TERMINATE to shut down."
			}
		}
	}()
}

/*
useTool converts the tool call paramters into a struct from JSON, and then calls the appropriate type, based
on the tool name.
*/
func (system *System) useTool(toolCall openai.ToolCall) {
	errnie.Trace()

	var args map[string]interface{}
	if errnie.Error(json.Unmarshal([]byte(toolCall.Function.Arguments), &args)) != nil {
		return
	}

	utils.BeautifyToolCall(toolCall, args)

	switch toolCall.Function.Name {
	case "worker":
		worker := NewWorker()
		worker.Initialize()
		worker.Run(system.ctx, system.ID, args)
	}
}
