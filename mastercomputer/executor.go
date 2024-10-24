package mastercomputer

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
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
	errnie.Trace()

	return &Executor{
		conversation: NewConversation(),
		sequencer:    sequencer,
		toolset:      NewToolset(),
	}
}

func (executor *Executor) Close() {
	errnie.Trace()

	if executor.cancel != nil {
		executor.cancel()
	}
}

func (executor *Executor) Do(worker *Worker) {
	errnie.Trace()

	// Preserve existing conversation context
	if executor.conversation.context == nil || len(executor.conversation.context) == 0 {
		executor.conversation.Update(openai.SystemMessage(worker.system))
		executor.conversation.Update(openai.UserMessage(worker.user + "\n\n" + viper.GetViper().GetString("ai.prompt.guidance")))
	}

	for {
		if params, err := executor.prepareParams(worker); errnie.Error(err) == nil {
			if response, err := executor.executeCompletion(params); errnie.Error(err) == nil {
				executor.processResponse(worker, response)
			}
		}
	}
}

func (executor *Executor) prepareParams(worker *Worker) (openai.ChatCompletionNewParams, error) {
	errnie.Trace()

	messages := executor.conversation.Truncate()

	tools := worker.toolset

	// Override the tools if the worker is in a discussion.
	if worker.discussion != nil {
		tools = executor.toolset.Assign("discussion")
	}

	params := openai.ChatCompletionNewParams{
		Messages:       openai.F(messages),
		Model:          openai.F(openai.ChatModelGPT4oMini),
		Temperature:    openai.Float(utils.ToFixed(worker.temperature, 1)),
		ResponseFormat: openai.F(executor.getResponseFormat()),
		Tools:          openai.F(tools),
		Store:          openai.F(true),
	}

	return params, nil
}

var semaphore = make(chan struct{}, 1)

func (executor *Executor) executeCompletion(params openai.ChatCompletionNewParams) (response *openai.ChatCompletion, err error) {
	errnie.Trace()

	semaphore <- struct{}{}
	defer func() { <-semaphore }()
	completion := NewCompletion(executor.ctx)
	return completion.Execute(executor.ctx, params)
}

func (executor *Executor) processResponse(worker *Worker, response *openai.ChatCompletion) {
	errnie.Trace()

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
		fmt.Println(content)
		wrappedContent := fmt.Sprintf("[%s]\n%s\n[/%s]\n", worker.name, content, worker.name)
		wrappedMessage := openai.ChatCompletionMessage{
			Role:    message.Role,
			Content: wrappedContent,
		}
		executor.conversation.Update(wrappedMessage)

		if worker.discussion != nil {
			worker.discussion.conversation.Update(wrappedMessage)
		}
	}

	toolMessages := executor.handleToolCalls(worker, response)
	for _, toolMessage := range toolMessages {
		executor.conversation.Update(toolMessage)

		if worker.discussion != nil {
			worker.discussion.conversation.Update(toolMessage)
		}
	}

	executor.handleDiscussion(worker)
}

func (executor *Executor) handleDiscussion(worker *Worker) {
	errnie.Trace()

	if worker.discussion != nil {
		count := 0
		for _, wrkr := range executor.sequencer.workers[worker.role] {
			if wrkr.state == WorkerStateAgreed {
				count++
			}
		}

		if count == len(executor.sequencer.workers[worker.role]) {
			for _, wrkr := range executor.sequencer.workers[worker.role] {
				wrkr.discussion = nil
			}
		}
	}
}

func (executor *Executor) handleToolCalls(worker *Worker, response *openai.ChatCompletion) []openai.ChatCompletionMessageParamUnion {
	errnie.Trace()

	executor.conversation.UpdateTokenCounts(response.Usage)
	message := response.Choices[0].Message

	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		return nil
	}

	results := []openai.ChatCompletionMessageParamUnion{}

	for _, toolCall := range message.ToolCalls {
		result := executor.toolset.Use(executor.sequencer, worker, toolCall)
		results = append(results, result)
	}

	if len(results) > 0 {
		results = append([]openai.ChatCompletionMessageParamUnion{message}, results...)
	}

	return results
}

func (executor *Executor) getResponseFormat() openai.ChatCompletionNewParamsResponseFormatUnion {
	return openai.ResponseFormatJSONSchemaParam{
		Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
		JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        openai.F("reasoning"),
			Description: openai.F("response format"),
			Schema: openai.F(interface{}(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"thoughts": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/definitions/thought",
						},
					},
					"done": map[string]interface{}{
						"type": "boolean",
					},
				},
				"required":             []string{"thoughts", "done"},
				"additionalProperties": false,
				"definitions": map[string]interface{}{
					"thought": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"chain": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/definitions/thought",
								},
							},
							"tree": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/definitions/thought",
								},
							},
							"ideas": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/definitions/thought",
								},
							},
							"realizations": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/definitions/thought",
								},
							},
							"self_reflections": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/definitions/thought",
								},
							},
						},
						"required":             []string{"chain", "tree", "ideas", "realizations", "self_reflections"},
						"additionalProperties": false,
					},
				},
			})),
			Strict: openai.F(true),
		}),
	}
}
