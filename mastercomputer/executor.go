package mastercomputer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

var maxContextTokens = 128000

type Executor struct {
	parentCtx    context.Context
	ctx          context.Context
	cancel       context.CancelFunc
	worker       *Worker
	conversation *Conversation
}

func NewExecutor(ctx context.Context, worker *Worker) *Executor {
	return &Executor{
		parentCtx:    ctx,
		worker:       worker,
		conversation: NewConversation(worker.buffer, maxContextTokens),
	}
}

func (executor *Executor) Initialize() error {
	executor.ctx, executor.cancel = context.WithCancel(context.Background())
	return nil
}

func (executor *Executor) Close() {
	if executor.cancel != nil {
		executor.cancel()
	}
}

func (executor *Executor) Execute(message *data.Artifact) {
	defer executor.Close()
	executor.conversation.Initialize()

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

func (executor *Executor) prepareParams() (openai.ChatCompletionNewParams, error) {
	messages := executor.conversation.Truncate()

	responseFormat, err := executor.getResponseFormat(executor.worker.buffer.Peek("workload"))
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	return openai.ChatCompletionNewParams{
		Messages:       openai.F(messages),
		ResponseFormat: openai.F(responseFormat),
		Tools: openai.F(NewToolset(
			executor.worker.buffer.Peek("workload")).tools,
		),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(0.0),
	}, nil
}

var semaphore = make(chan struct{}, 1)

func (executor *Executor) executeCompletion(params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	semaphore <- struct{}{}        // Acquire a token
	defer func() { <-semaphore }() // Release the token

	completion := NewCompletion(executor.ctx)
	response, err := completion.Execute(executor.ctx, params)
	if err != nil {
		var apiError *openai.Error
		if errors.As(err, &apiError) {
			switch apiError.StatusCode {
			case 429:
				log.Printf("Rate limit exceeded: %v", apiError)
				time.Sleep(time.Minute)
				return nil, apiError
			case 401:
				log.Printf("Authentication error: %v", apiError)
				return nil, apiError
			default:
				log.Printf("OpenAI API error: %v", apiError)
				return nil, apiError
			}
		} else {
			log.Printf("Unexpected error: %v", err)
			return nil, err
		}
	}

	return response, nil
}

func (executor *Executor) processResponse(response *openai.ChatCompletion) {
	if response == nil || len(response.Choices) == 0 {
		log.Println("No response from OpenAI")
		return
	}

	if response.Usage.CompletionTokens > 0 {
		executor.conversation.UpdateTokenCounts(response.Usage)
	}

	message := response.Choices[0].Message
	executor.conversation.Update(message)

	content := message.Content
	if content == "" {
		errnie.Warn("Empty response content")
		return
	}

	if err := executor.printResponse(content); err != nil {
		errnie.Error(err)
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
		errnie.Info("<tool call> %s", toolCall.Function.Name)
		result, err := UseTool(toolCall)

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
	return openai.ResponseFormatJSONSchemaParam{
		Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
		JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        openai.F(workload),
			Description: openai.F(workload + " format"),
			Schema:      openai.F(GenerateSchema[format.Reasoning]()),
		}),
	}, nil
}
