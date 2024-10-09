package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
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
	I        chan string
	O        chan string
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
		I:        make(chan string),
		O:        make(chan string),
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

func (worker *Worker) Initialize() error {
	errnie.Trace()
	return nil
}

func (worker *Worker) Run(ctx context.Context, parentID string, args map[string]any) (string, error) {
	errnie.Trace()

	worker.parentID = parentID
	worker.system = args["system"].(string)
	worker.user = args["user"].(string)
	worker.toolset = args["toolset"].(string)

	worker.Generate()

	return "", nil
}

func (worker *Worker) Generate() {
	errnie.Trace()

	broadcast := worker.queue.Register(worker.ID)

	go func() {
		for {
			select {
			case <-worker.ctx.Done():
				return
			case message := <-broadcast:
				worker.messages = append(worker.messages, message)
			case prompt := <-worker.I:
				errnie.Info("%s received prompt: %s", worker.ID, prompt)
				if worker.parentID != "" {
					worker.queue.Publish(*data.New(
						worker.ID,
						"ACK",
						worker.parentID,
						[]byte("READY! Received: "+prompt),
					))
				}

				worker.Workload = append(worker.Workload, prompt)
			default:
				if len(worker.Workload) == 0 {
					time.Sleep(1 * time.Second)
					continue
				}

				// Step 1: Analyze Prompt and Choose Reasoning Strategies
				var prompt string
				prompt, worker.Workload = worker.Workload[0], worker.Workload[1:]
				prompt = worker.processMessages(prompt) + "\n\n" + worker.Memory.ToString()

				initialMessage := worker.printResponse(NewCompletion(worker.ctx).Execute(
					worker.system,
					prompt+"\nPlease begin by analyzing the prompt and selecting one or more appropriate reasoning strategies.",
					worker.toolset,
					format.NewReasoningStrategy(),
				))

				// Step 2: Extract the selected strategies
				strategies := worker.ExtractReasoningStrategy(initialMessage)

				// Step 3: Loop over the selected strategies
				for _, strategy := range strategies {
					strategyResponse := worker.printResponse(NewCompletion(worker.ctx).Execute(
						worker.system,
						prompt+"\nProceed with the selected strategy: "+strategy.Name(),
						worker.toolset,
						strategy,
					))

					// Pretty-print the current reasoning strategy response
					utils.BeautifyReasoning(worker.ID, strategy)

					// Update worker memory with the response
					worker.Memory.ShortTerm = append(worker.Memory.ShortTerm, strategyResponse.Content)

					// Pretty-print memory after every reasoning step
					utils.BeautifyMemory(worker.Memory)
				}

				// Step 4: Verification Step
				verificationPrompt := "\nCombined Responses:\n" + worker.Memory.ToString() + "\nPlease verify correctness, coherence, and completeness of each response."
				verificationResponse := worker.printResponse(NewCompletion(worker.ctx).Execute(
					worker.system,
					verificationPrompt,
					worker.toolset,
					format.NewFinal(),
				))

				// Pretty-print verification response
				fmt.Println("Verification Response:")
				fmt.Println(color.MagentaString(verificationResponse.Content))

				// Step 5: Generate Final Answer
				finalPrompt := prompt + "\n\nVerification Results:\n" + verificationResponse.Content + "\n\nBased on the verified reasoning, please provide a final answer."
				finalMessage := worker.printResponse(NewCompletion(worker.ctx).Execute(
					worker.system,
					finalPrompt,
					worker.toolset,
					format.NewFinal(),
				))

				// Publish the final response if there's a parent system to send to.
				if worker.parentID != "" {
					worker.queue.Publish(*data.New(worker.ID, worker.parentID, finalMessage.Content, []byte(finalMessage.Content)))
				}
			}
		}
	}()
}

func (worker *Worker) ExtractReasoningStrategy(message openai.ChatCompletionMessage) []format.Response {
	errnie.Trace()

	var strategy format.ReasoningStrategy

	if errnie.Error(json.Unmarshal([]byte(message.Content), &strategy)) != nil {
		return []format.Response{
			format.NewChainOfThought(),
		}
	}

	var out []format.Response

	for _, strategy := range strategy.Template.Strategies {
		switch strategy.Strategy {
		case format.StrategyChainOfThought:
			out = append(out, format.NewChainOfThought())
		case format.StrategyTreeOfThought:
			out = append(out, format.NewTreeOfThought())
		case format.StrategyFirstPrinciples:
			out = append(out, format.NewFirstPrinciplesReasoning())
		}
	}

	return out
}

func (worker *Worker) printResponse(response openai.ChatCompletionResponse) openai.ChatCompletionMessage {
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

		var chainOfThought format.ChainOfThought

		if errnie.Error(json.Unmarshal([]byte(message.Content), &chainOfThought.Template)) != nil {
			return message
		}

		utils.BeautifyChainOfThought(worker.ID, chainOfThought)

		if strings.ToUpper(chainOfThought.Template.Action) == "TERMINATE" {
			return message
		}
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

	switch toolCall.Function.Name {
	case "worker":
		worker := NewWorker()
		worker.Initialize()
		worker.Run(worker.ctx, worker.ID, args)
	}
}
