package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

type Worker struct {
	ctx      context.Context
	ID       string
	parentID string
	system   string
	user     string
	toolset  string
	Memory   *ai.Memory
	Workload []string
	queue    *twoface.Queue
	messages []data.Artifact
	Function *openai.FunctionDefinition
}

func NewWorker() *Worker {
	errnie.Trace()

	return &Worker{
		ctx:      context.Background(),
		ID:       utils.NewID(),
		messages: make([]data.Artifact, 0),
		Memory:   ai.NewMemory(),
		Workload: make([]string, 0),
		queue:    twoface.NewQueue(),
		Function: &openai.FunctionDefinition{
			Name:        "worker",
			Description: "Use to create a worker agent, which can become anything you can imagine, using the system and user prompt, and providing a toolset.",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Description:          "Use to create a worker agent, pass in the system prompt and user prompt",
				Properties: map[string]jsonschema.Definition{
					"system": {
						Type:        jsonschema.String,
						Description: "The system prompt",
					},
					"user": {
						Type:        jsonschema.String,
						Description: "The user prompt",
					},
					"toolset": {
						Type:        jsonschema.String,
						Description: "The toolset to use, or 'none' if no toolset is required",
					},
				},
				Required: []string{"system", "user", "toolset"},
			},
		},
	}
}

func (worker *Worker) Initialize() *Worker {
	errnie.Trace()
	return worker
}

func (worker *Worker) Run(ctx context.Context, prompt *data.Artifact) *data.Artifact {
	errnie.Trace()

	var (
		payload []byte
		err     error
	)

	payload, err = prompt.Payload()

	if err != nil {
		return nil
	}

	loaded := worker.processMessages(string(payload)) + "\n\n" + worker.Memory.ToString()

	response := worker.printResponse(NewCompletion(worker.ctx).Execute(
		worker.system,
		worker.user+"\n\n"+loaded+"\n\nPlease start by selecting one or more reasoning strategies. Available strategies are: chain_of_thought, tree_of_thought, first_principles",
		worker.toolset,
		format.NewReasoningStrategy(),
	), "reasoning_strategy")

	for _, strategy := range worker.ExtractReasoningStrategy(response) {
		errnie.Info("applying strategy: %s", strategy)

		response = worker.printResponse(NewCompletion(worker.ctx).Execute(
			worker.system,
			worker.user+"\n\n"+loaded,
			worker.toolset,
			strategymap[strategy],
		), strategy)
	}

	return data.New(worker.ID, "response", "broadcast", []byte(response.Content))
}

var strategymap = map[string]format.Response{
	"reasoning_strategy": format.NewReasoningStrategy(),
	"chain_of_thought":   format.NewChainOfThought(),
	"tree_of_thought":    format.NewTreeOfThought(),
	"first_principles":   format.NewFirstPrinciplesReasoning(),
}

func (worker *Worker) ExtractReasoningStrategy(response openai.ChatCompletionMessage) []string {
	errnie.Trace()
	buf := format.ReasoningStrategy{}.Template
	errnie.Error(json.Unmarshal([]byte(response.Content), &buf))
	return buf.OrderedStrategies
}

func (worker *Worker) printResponse(response openai.ChatCompletionResponse, strategy string) openai.ChatCompletionMessage {
	errnie.Trace()

	message := response.Choices[0].Message

	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			worker.Memory.ShortTerm = append(worker.Memory.ShortTerm, fmt.Sprintf("TOOL CALL: %s, ARGUMENTS: %s", toolCall.Function.Name, toolCall.Function.Arguments))
			worker.useTool(toolCall)
		}
	}

	if message.Content != "" {
		worker.Memory.ShortTerm = append(worker.Memory.ShortTerm, message.Content)
		utils.BeautifyReasoning(worker.ID, strategymap[strategy])
	}

	return message
}

func (worker *Worker) processMessages(prompt string) string {
	errnie.Trace()

	// Add messages to the prompt context.
	prompt += "\n\n[MESSAGES]\n"
	for _, message := range worker.messages {
		if origin, err := message.Origin(); err == nil {
			if role, err := message.Role(); err == nil {
				if scope, err := message.Scope(); err == nil {
					prompt += fmt.Sprintf("[%s @ %s]: %s, %s\n", origin, time.Now().Format("15:04:05"), role, scope)
				}
			}
		}
	}
	prompt += "\n[/MESSAGES]\n\n"

	return prompt
}

/*
useTool converts the tool call paramters into a struct from JSON, and then calls the appropriate type, based
on the tool name.
*/
func (worker *Worker) useTool(toolCall openai.ToolCall) {
	errnie.Trace()

	var args map[string]interface{}
	if errnie.Error(json.Unmarshal([]byte(toolCall.Function.Arguments), &args)) != nil {
		return
	}

	utils.BeautifyToolCall(toolCall, args)

	prompt := data.New(
		worker.ID,
		"prompt",
		"broadcast",
		[]byte(args["user"].(string)),
	)

	switch toolCall.Function.Name {
	case "worker":
		worker := NewWorker()
		worker.Initialize()
		worker.Run(worker.ctx, prompt)
	}
}
