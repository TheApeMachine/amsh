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
	"github.com/theapemachine/amsh/utils"
)

var maxContextTokens = 128000

type Executor struct {
	parentCtx    context.Context
	ctx          context.Context
	cancel       context.CancelFunc
	task         *data.Artifact
	conversation *Conversation
}

func NewExecutor(pctx context.Context, task *data.Artifact) *Executor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Executor{
		parentCtx:    pctx,
		ctx:          ctx,
		cancel:       cancel,
		task:         task,
		conversation: NewConversation(),
	}
}

func (executor *Executor) Close() {
	if executor.cancel != nil {
		executor.cancel()
	}
}

func (executor *Executor) Do() *data.Artifact {
	defer executor.Close()

	params, err := executor.prepareParams()
	if err != nil {
		log.Printf("Error preparing parameters: %v", err)
		return
	}

	response, err := executor.executeCompletion(params)
	if err != nil {
		log.Printf("Error executing completion: %v", err)
		return
	}

	executor.processResponse(response)
}

func (executor *Executor) Verify() {
	errnie.Info("verifying")
	verification := data.New(executor.worker.name, executor.worker.buffer.Peek("workload"), "verifying", []byte{})
	executor.worker.queue.Publish(verification)
}

func (executor *Executor) prepareParams() (openai.ChatCompletionNewParams, error) {
	messages := executor.conversation.Truncate()

	responseFormat, err := executor.getResponseFormat(executor.worker.buffer.Peek("workload"))
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	// Convert the string to a float
	temperature, err := strconv.ParseFloat(executor.worker.buffer.Peek("temperature"), 64)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	errnie.Note("%s generating with temperature: %f", executor.worker.name, temperature)

	return openai.ChatCompletionNewParams{
		Messages:       openai.F(messages),
		ResponseFormat: openai.F(responseFormat),
		Tools: openai.F(NewToolset(
			executor.worker.buffer.Peek("workload")).tools,
		),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(utils.ToFixed(temperature, 1)),
		Store:       openai.F(true),
	}, nil
}

var semaphore = make(chan struct{}, 3)

func (executor *Executor) executeCompletion(
	params openai.ChatCompletionNewParams,
) (response *openai.ChatCompletion, err error) {
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
	executor.buffer.Write([]byte(message.Content))

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

	errnie.Info("worker: %s", executor.worker.name)
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
		result, err := UseTool(toolCall, executor.worker)

		if err != nil {
			errnie.Error(err)
			return nil
		}

		results = append(results, openai.ToolMessage(toolCall.ID, result))
	}

	return results
}

func (executor *Executor) getResponseFormat(workload string) (
	openai.ChatCompletionNewParamsResponseFormatUnion, error,
) {
	schema := GenerateSchema[format.Reasoning]()

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
