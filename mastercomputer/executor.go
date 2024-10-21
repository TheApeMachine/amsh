package mastercomputer

import (
	"context"
	"log"
	"strconv"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

type Strategy interface {
	String() string
}

type Executor struct {
	ctx          context.Context
	cancel       context.CancelFunc
	queue        *twoface.Queue
	worker       *Worker
	toolset      *Toolset
	task         *data.Artifact
	conversation *Conversation
	strategy     Strategy
	eventEmitter *twoface.EventEmitter
}

func NewExecutor(worker *Worker, task *data.Artifact) *Executor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Executor{
		ctx:          ctx,
		cancel:       cancel,
		queue:        twoface.NewQueue(),
		worker:       worker,
		toolset:      NewToolset(),
		task:         task,
		conversation: NewConversation(task),
		eventEmitter: twoface.NewEventEmitter(),
	}
}

func (executor *Executor) Close() {
	if executor.cancel != nil {
		executor.cancel()
	}
}

func (executor *Executor) Do(maxIterations int) bool {
	executor.worker.state = WorkerStateBusy

	defer executor.Close()

	iteration := 0

	for {
		iteration++
		isDone := false

		if params, err := executor.prepareParams(iteration, maxIterations); errnie.Error(err) == nil {
			if response, err := executor.executeCompletion(params); errnie.Error(err) == nil {
				isDone = executor.processResponse(response)
			}
		}

		if isDone {
			break
		}

		errnie.Debug("\n\n---PAYLOAD---\n\n%s\n\n---/PAYLOAD--\n\n", executor.task.Peek("payload"))
	}

	return true
}

func (executor *Executor) prepareParams(iteration, maxIterations int) (openai.ChatCompletionNewParams, error) {
	messages := executor.conversation.Truncate()
	iterationMsg := "\n\n[ITERATION: " + strconv.Itoa(iteration) + " of " + strconv.Itoa(maxIterations) + "]\n\n"

	additionalReminder := "\n\nBe mindful of your current iteration count!\n"
	if iteration == maxIterations-1 {
		additionalReminder = "You are getting close to the end of your permitted iterations, make sure to wrap up any loose ends!\n\n"
	}

	executor.task.Append(iterationMsg + additionalReminder)
	messages = append(messages, openai.AssistantMessage(iterationMsg))

	responseFormat, err := executor.getResponseFormat(executor.worker.buffer.Peek("role"))
	if errnie.Error(err) != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	if executor.worker.buffer.Peek("role") == "reflection" {
		responseFormat, err = executor.getResponseFormat("reflection")
		if errnie.Error(err) != nil {
			return openai.ChatCompletionNewParams{}, err
		}
	}

	// Convert the string to a float
	temperature, err := strconv.ParseFloat(executor.worker.buffer.Peek("temperature"), 64)
	if errnie.Error(err) != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	errnie.Note(
		"%s (%s) generating with temperature: %.1f [ITERATION: %d of %d]",
		executor.worker.name,
		executor.worker.buffer.Peek("role"),
		temperature,
		iteration,
		viper.GetViper().GetInt("ai.max_iterations"),
	)

	params := openai.ChatCompletionNewParams{
		Messages:       openai.F(messages),
		ResponseFormat: openai.F(responseFormat),
		Model:          openai.F(openai.ChatModelGPT4oMini),
		Temperature:    openai.Float(utils.ToFixed(temperature, 1)),
		Store:          openai.F(true),
	}

	// Make sure it doesn't just straight to tool calling, but always starts with reasoning.
	if iteration > 1 {
		params.Tools = openai.F(executor.toolset.Assign(executor.worker.buffer.Peek("role")))
	}

	return params, nil
}

var semaphore = make(chan struct{}, 1)

func (executor *Executor) executeCompletion(params openai.ChatCompletionNewParams) (response *openai.ChatCompletion, err error) {
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	completion := NewCompletion(executor.ctx)
	return completion.Execute(executor.ctx, params)
}

func (executor *Executor) processResponse(response *openai.ChatCompletion) (isDone bool) {
	var err error

	if len(response.Choices) == 0 {
		log.Println("No response from OpenAI")
		return
	}

	if response.Usage.CompletionTokens > 0 {
		executor.conversation.UpdateTokenCounts(response.Usage)
	}

	message := response.Choices[0].Message

	content := message.Content
	if content != "" {
		executor.conversation.Update(message)
		executor.task.Append(message.Content)

		errnie.Debug("Emitting executor update event")
		executor.EmitEvent(twoface.EventTypeExecutorUpdate, *executor.task)

		if isDone, err = executor.printResponse(content); err != nil {
			errnie.Error(err)
		} else if isDone {
			return true
		}
	}

	toolMessages := executor.handleToolCalls(response)
	for _, toolMessage := range toolMessages {
		executor.conversation.Update(toolMessage)
		switch msg := toolMessage.(type) {
		case openai.ChatCompletionMessage:
			executor.task.Append(msg.Content)
		case openai.ChatCompletionToolMessageParam:
			executor.task.Append(msg.Content.String())
		}
	}

	return false
}

func (executor *Executor) printResponse(content string) (isDone bool, err error) {
	switch executor.strategy.(type) {
	case format.Reasoning:
		return format.NewReasoning().Print([]byte(content))
	case format.Working:
		return format.NewWorking().Print([]byte(content))
	case format.Reviewing:
		return format.NewReviewing().Print([]byte(content))
	case format.Verifying:
		return format.NewVerifying().Print([]byte(content))
	case format.Executing:
		return format.NewExecuting().Print([]byte(content))
	case format.Communicating:
		return format.NewCommunicating().Print([]byte(content))
	case format.Managing:
		return format.NewManaging().Print([]byte(content))
	case format.SelfReflection:
		return format.NewSelfReflection().Print([]byte(content))
	}

	return false, nil
}

func (executor *Executor) handleToolCalls(response *openai.ChatCompletion) []openai.ChatCompletionMessageParamUnion {
	message := response.Choices[0].Message

	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		return nil
	}

	results := []openai.ChatCompletionMessageParamUnion{}

	for _, toolCall := range message.ToolCalls {
		errnie.Note("TOOL CALL: %s - %v", toolCall.Function.Name, toolCall.Function.Arguments)
		result := executor.toolset.Use(executor.task.Peek("origin"), toolCall)

		errnie.Note("[TOOL RESULT %s]\n%s\n[/TOOL RESULT]\n", toolCall.ID, result.Content.String())

		results = append(results, result)
	}

	if len(results) > 0 {
		results = append([]openai.ChatCompletionMessageParamUnion{message}, results...)
	}

	return results
}

func (executor *Executor) getResponseFormat(workload string) (
	openai.ChatCompletionNewParamsResponseFormatUnion, error,
) {
	var schema interface{}

	switch workload {
	case "reasoner":
		executor.strategy = format.Reasoning{}
		schema = GenerateSchema[format.Reasoning]()
	case "researcher":
		executor.strategy = format.Working{}
		schema = GenerateSchema[format.Working]()
	case "reviewer":
		executor.strategy = format.Reviewing{}
		schema = GenerateSchema[format.Reviewing]()
	case "verifier":
		executor.strategy = format.Verifying{}
		schema = GenerateSchema[format.Verifying]()
	case "executor":
		executor.strategy = format.Executing{}
		schema = GenerateSchema[format.Executing]()
	case "communicator":
		executor.strategy = format.Communicating{}
		schema = GenerateSchema[format.Communicating]()
	case "manager":
		executor.strategy = format.Managing{}
		schema = GenerateSchema[format.Managing]()
	case "reflection":
		executor.strategy = format.SelfReflection{}
		schema = GenerateSchema[format.SelfReflection]()
	default:
		executor.strategy = format.Working{}
		schema = GenerateSchema[format.Working]()
	}

	return openai.ResponseFormatJSONSchemaParam{
		Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
		JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        openai.F(workload),
			Description: openai.F("response format"),
			Schema:      openai.F(schema),
			Strict:      openai.F(true),
		}),
	}, nil
}

// Add this method to the Executor struct
func (executor *Executor) EmitEvent(eventType twoface.EventType, payload data.Artifact) {
	executor.eventEmitter.Emit(twoface.Event{Type: eventType, Payload: payload})
}
