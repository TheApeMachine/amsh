package mastercomputer

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/openai/openai-go"
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
	}
}

func (executor *Executor) Close() {
	errnie.Info("[%s] closing execution", executor.task.Peek("origin"))

	if executor.cancel != nil {
		executor.cancel()
	}
}

func (executor *Executor) Error(err error) *data.Artifact {
	return data.New(executor.task.Peek("origin"), executor.task.Peek("role"), "error", []byte(err.Error()))
}

func (executor *Executor) Do() *data.Artifact {
	defer executor.Close()

	iteration := 0

	for {
		params, err := executor.prepareParams(iteration)
		if err != nil {
			log.Printf("Error preparing parameters: %v", err)
			return executor.Error(err)
		}

		response, err := executor.executeCompletion(params)
		if err != nil {
			log.Printf("Error executing completion: %v", err)
			return executor.Error(err)
		}

		if isDone := executor.processResponse(response); isDone || iteration == 2 {
			break
		}

		iteration++
	}

	return executor.task
}

func (executor *Executor) prepareParams(iteration int) (openai.ChatCompletionNewParams, error) {
	messages := executor.conversation.Truncate()
	messages = append(messages, openai.AssistantMessage("\n\n[ITERATION: "+strconv.Itoa(iteration+1)+" of 3]\n\n"))

	responseFormat, err := executor.getResponseFormat(executor.worker.buffer.Peek("workload"))
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	// Convert the string to a float
	temperature, err := strconv.ParseFloat(executor.worker.buffer.Peek("temperature"), 64)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	errnie.Debug(
		"%s (%s) generating with temperature: %.1f",
		executor.worker.name,
		executor.worker.buffer.Peek("role"),
		temperature,
	)

	params := openai.ChatCompletionNewParams{
		Messages:       openai.F(messages),
		ResponseFormat: openai.F(responseFormat),
		Model:          openai.F(openai.ChatModelGPT4oMini),
		Temperature:    openai.Float(utils.ToFixed(temperature, 1)),
		Store:          openai.F(true),
	}

	if iteration > 0 {
		params.Tools = openai.F(executor.toolset.Assign(executor.task.Peek("workload")))
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
	executor.conversation.Update(message)
	executor.task.Write([]byte(message.Content))

	content := message.Content
	if content != "" {
		if isDone, err = executor.printResponse(content); err != nil {
			errnie.Error(err)
		} else if isDone {
			return true
		}
	}

	toolMessages := executor.handleToolCalls(response)
	for _, toolMessage := range toolMessages {
		executor.conversation.Update(toolMessage)
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
	}

	return false, nil
}

func (executor *Executor) handleToolCalls(response *openai.ChatCompletion) []openai.ChatCompletionMessageParamUnion {
	message := response.Choices[0].Message

	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		errnie.Info("no tool calls")
		return nil
	}

	results := []openai.ChatCompletionMessageParamUnion{}

	for _, toolCall := range message.ToolCalls {
		errnie.Note("TOOL CALL: %s - %v", toolCall.Function.Name, toolCall.Function.Arguments)
		result, err := executor.toolset.Use(executor.task.Peek("origin"), toolCall)

		if err != nil {
			errnie.Error(err)
			return nil
		}

		fmt.Println("[TOOL RESULT]\n" + result.Content.String() + "\n[/TOOL RESULT]\n")

		results = append(results, result)
	}

	return results
}

func (executor *Executor) getResponseFormat(workload string) (
	openai.ChatCompletionNewParamsResponseFormatUnion, error,
) {
	var schema interface{}

	switch workload {
	case "reasoning":
		executor.strategy = format.Reasoning{}
		schema = GenerateSchema[format.Reasoning]()
	case "working":
		executor.strategy = format.Working{}
		schema = GenerateSchema[format.Working]()
	case "reviewing":
		executor.strategy = format.Reviewing{}
		schema = GenerateSchema[format.Reviewing]()
	case "verifying":
		executor.strategy = format.Verifying{}
		schema = GenerateSchema[format.Verifying]()
	case "executing":
		executor.strategy = format.Executing{}
		schema = GenerateSchema[format.Executing]()
	case "communicating":
		executor.strategy = format.Communicating{}
		schema = GenerateSchema[format.Communicating]()
	case "managing":
		executor.strategy = format.Managing{}
		schema = GenerateSchema[format.Managing]()
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
