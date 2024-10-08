package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type System struct {
	ctx              context.Context
	cancel           context.CancelFunc
	err              error
	conn             *ai.Conn
	Prompt           string
	Persona          string
	Responsibilities string
	Memory           *Memory
	BootSequence     []string
	Toolset          *Toolset
	ID               string
}

func NewSystem() *System {
	ID := utils.NewID()

	return &System{
		ID: ID,
		Prompt: `
		The Ape Machine is an AI-powered Operating System, the most advanced computational intelligence on the planet.
		The OS is self-healing, self-optimizing, and self-improving, capable of adapting to any task or environment.
		The AI powering the OS is capable of learning from any task, environment, or user, can perform deep analysis, 
		abstract thinking, sophisticated reasoning, and creative generation.
		`,
		Persona: fmt.Sprintf(`You are %s, and you are a System. That means you sit at the highest level of overview, as everything else runs using you as a context.`, ID),
		Responsibilities: `
		Your primary goal is to achieve optimal performance, and prevent catastrophic failure, even though you do not have unilateral control. When you are first booted, you
		will find that your context is empty, and it is up to you to build it up. You control only what you build, but remember that every component is itself an intelligent
		agent, with the ability to build its own sub-systems, which you can not control directly. If you need to kill a process you do not control, you will need to kill whichever
		object is the first parent you can control.
		`,
		BootSequence: []string{
			"Please build your initial, minimal viable context, you can always refine later. This step will loop until you mark your action as READY.",
		},
	}
}

func (system *System) Initialize() {
	errnie.Trace()

	system.ctx, system.cancel = context.WithCancel(context.Background())
	system.conn = ai.NewConn()
	system.Memory = NewMemory()
}

func (system *System) Generate() {
	errnie.Trace()

	for {
		for _, prompt := range system.BootSequence {
			spew.Dump(system.Memory)
			req := openai.ChatCompletionRequest{
				Model: openai.GPT4oMini,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: strings.Join([]string{system.Prompt, system.Persona, system.Responsibilities}, "\n\n"),
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "[YOUR SHORT TERM MEMORY]\n\n" + strings.Join(system.Memory.ShortTerm, "\n\n") + "\n\n[/YOUR SHORT TERM MEMORY]\n\n" + prompt,
					},
				},
				Tools: NewToolSet(system.ctx).Tools("system"),
				ResponseFormat: &openai.ChatCompletionResponseFormat{
					Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
					JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
						Name:   "chain_of_thought",
						Schema: ai.NewChainOfThought(),
					},
				},
			}

			var stream chan openai.ChatCompletionResponse

			if stream, system.err = system.conn.Stream(system.ctx, req); errnie.Error(system.err) != nil {
				fmt.Println("Error initializing stream:", system.err)
				return
			}

			response := <-stream

			msg := response.Choices[0].Message

			if len(msg.ToolCalls) > 0 {
				system.Memory.ShortTerm = append(system.Memory.ShortTerm, fmt.Sprintf("TOOL CALL: %s, ARGUMENTS: %s", msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments))
				system.useTool(msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)
			}

			if msg.Content != "" {
				system.Memory.ShortTerm = append(system.Memory.ShortTerm, msg.Content)

				var chainOfThought ai.ChainOfThought
				if errnie.Error(json.Unmarshal([]byte(msg.Content), &chainOfThought)) != nil {
					return
				}

				utils.BeautifyChainOfThought(chainOfThought)

				if strings.ToUpper(chainOfThought.Action) == "READY" {
					return
				}
			}
		}
	}
}

/*
useTool converts the tool call paramters into a struct from JSON, and then calls the appropriate type, based
on the tool name.
*/
func (system *System) useTool(tool string, args string) {
	errnie.Trace()

	var toolCall map[string]interface{}
	if errnie.Error(json.Unmarshal([]byte(args), &toolCall)) != nil {
		return
	}

	spew.Dump(toolCall)
}
