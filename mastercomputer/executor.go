package mastercomputer

import (
	"context"
	"log"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

type Executor struct {
	ctx          context.Context
	cancel       context.CancelFunc
	sequencer    *Sequencer
	conversation *Conversation
	toolset      *Toolset
}

func NewExecutor(sequencer *Sequencer) *Executor {
	return &Executor{
		conversation: NewConversation(),
		sequencer:    sequencer,
		toolset:      NewToolset(),
	}
}

func (executor *Executor) Close() {
	if executor.cancel != nil {
		executor.cancel()
	}
}

func (executor *Executor) Do(worker *Worker) {
	if params, err := executor.prepareParams(worker); errnie.Error(err) == nil {
		if response, err := executor.executeCompletion(params); errnie.Error(err) == nil {
			executor.processResponse(worker, response)
		}
	}
}

func (executor *Executor) prepareParams(worker *Worker) (openai.ChatCompletionNewParams, error) {
	messages := executor.conversation.Truncate()

	params := openai.ChatCompletionNewParams{
		Messages:    openai.F(messages),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(utils.ToFixed(worker.temperature, 1)),
		Store:       openai.F(true),
	}

	if worker.format != nil {
		params.ResponseFormat = openai.F(worker.format)
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

func (executor *Executor) processResponse(worker *Worker, response *openai.ChatCompletion) {
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
	}

	toolMessages := executor.handleToolCalls(worker, response)
	for _, toolMessage := range toolMessages {
		executor.conversation.Update(toolMessage)
	}
}

func (executor *Executor) handleToolCalls(worker *Worker, response *openai.ChatCompletion) []openai.ChatCompletionMessageParamUnion {
	executor.conversation.UpdateTokenCounts(response.Usage)
	message := response.Choices[0].Message

	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		return nil
	}

	results := []openai.ChatCompletionMessageParamUnion{}

	for _, toolCall := range message.ToolCalls {
		result := executor.toolset.Use(executor.sequencer, toolCall)
		results = append(results, result)
	}

	if len(results) > 0 {
		results = append([]openai.ChatCompletionMessageParamUnion{message}, results...)
	}

	return results
}

func (executor *Executor) getResponseFormat(worker *Worker) openai.ChatCompletionNewParamsResponseFormatUnion {
	return openai.ResponseFormatJSONSchemaParam{
		Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
		JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        openai.F(worker.name),
			Description: openai.F("response format"),
			Schema:      openai.F(worker.schema),
			Strict:      openai.F(true),
		}),
	}
}
