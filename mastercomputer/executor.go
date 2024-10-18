package mastercomputer

import (
	"context"
	"encoding/json"
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

type Executor struct {
	parentCtx    context.Context
	ctx          context.Context
	cancel       context.CancelFunc
	queue        *twoface.Queue
	toolset      *Toolset
	task         *data.Artifact
	conversation *Conversation
}

func NewExecutor(pctx context.Context, task *data.Artifact) *Executor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Executor{
		parentCtx:    pctx,
		ctx:          ctx,
		cancel:       cancel,
		queue:        twoface.NewQueue(),
		toolset:      NewToolset(),
		task:         task,
		conversation: NewConversation(),
	}
}

func (executor *Executor) Close() {
	if executor.cancel != nil {
		executor.cancel()
	}
}

func (executor *Executor) Error(err error) *data.Artifact {
	return data.New(executor.task.Peek("origin"), executor.task.Peek("role"), "error", []byte(err.Error()))
}

func (executor *Executor) Do() *data.Artifact {
	defer executor.Close()

	params, err := executor.prepareParams()
	if err != nil {
		log.Printf("Error preparing parameters: %v", err)
		return executor.Error(err)
	}

	response, err := executor.executeCompletion(params)
	if err != nil {
		log.Printf("Error executing completion: %v", err)
		return executor.Error(err)
	}

	executor.processResponse(response)

	executor.task.Poke("scope", "verifying")
	return executor.task
}

func (executor *Executor) prepareParams() (openai.ChatCompletionNewParams, error) {
	messages := executor.conversation.Truncate()

	responseFormat, err := executor.getResponseFormat(executor.task.Peek("workload"))
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	// Convert the string to a float
	temperature, err := strconv.ParseFloat(executor.task.Peek("temperature"), 64)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	errnie.Note("%s generating with temperature: %f", executor.task.Peek("origin"), temperature)

	return openai.ChatCompletionNewParams{
		Messages:       openai.F(messages),
		ResponseFormat: openai.F(responseFormat),
		Tools:          openai.F(executor.toolset.Assign(executor.task.Peek("workload"))),
		Model:          openai.F(openai.ChatModelGPT4oMini),
		Temperature:    openai.Float(utils.ToFixed(temperature, 1)),
		Store:          openai.F(true),
	}, nil
}

var semaphore = make(chan struct{}, 3)

func (executor *Executor) executeCompletion(params openai.ChatCompletionNewParams) (response *openai.ChatCompletion, err error) {
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	completion := NewCompletion(executor.ctx)
	return completion.Execute(executor.ctx, params)
}

func (executor *Executor) processResponse(response *openai.ChatCompletion) {
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
		errnie.Warn("Empty response content")
		if err := executor.printResponse(content); err != nil {
			errnie.Error(err)
		}
	}

	toolMessages := executor.handleToolCalls(response)
	for _, toolMessage := range toolMessages {
		executor.conversation.Update(toolMessage)
	}
}

func (executor *Executor) printResponse(content string) error {
	var strategy format.Reasoning

	if err := errnie.Error(json.Unmarshal([]byte(content), &strategy)); err != nil {
		return err
	}

	errnie.Info("worker: %s", executor.task.Peek("origin"))
	fmt.Println(strategy.String())
	return nil
}

func (executor *Executor) handleToolCalls(response *openai.ChatCompletion) []openai.ChatCompletionMessageParamUnion {
	message := response.Choices[0].Message

	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		errnie.Info("no tool calls")
		return nil
	}

	results := []openai.ChatCompletionMessageParamUnion{}

	for _, toolCall := range message.ToolCalls {
		errnie.Info("TOOL CALL: %s", toolCall.Function.Name)
		result, err := executor.toolset.Use(toolCall)

		if err != nil {
			errnie.Error(err)
			return nil
		}

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
		schema = GenerateSchema[format.Reasoning]()
	case "working":
		schema = GenerateSchema[format.Working]()
	case "reviewing":
		schema = GenerateSchema[format.Reviewing]()
	case "verifying":
		schema = GenerateSchema[format.Verifying]()
	case "executing":
		schema = GenerateSchema[format.Executing]()
	case "communicating":
		schema = GenerateSchema[format.Communicating]()
	case "managing":
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
