package provider

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go"
)

type OpenAI struct {
	client    *openai.Client
	model     string
	maxTokens int
}

func NewOpenAI(apiKey string, model string) *OpenAI {
	return &OpenAI{
		client:    openai.NewClient(),
		model:     model,
		maxTokens: 2000,
	}
}

func (o *OpenAI) GenerateStream(ctx context.Context, messages []Message) <-chan string {
	eventChan := make(chan string)
	go func() {
		defer close(eventChan)
		for event := range o.generateEvents(ctx, messages) {
			eventChan <- event.Content
		}
	}()
	return eventChan
}

func (o *OpenAI) generateEvents(ctx context.Context, messages []Message) <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)

		// Convert messages to OpenAI format
		openAIMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
		for i, msg := range messages {
			switch msg.Role {
			case "user":
				openAIMessages[i] = openai.UserMessage(msg.Content)
			case "assistant":
				openAIMessages[i] = openai.AssistantMessage(msg.Content)
			case "system":
				openAIMessages[i] = openai.SystemMessage(msg.Content)
			case "tool":
				openAIMessages[i] = openai.ToolMessage(msg.Name, msg.Content)
			}
		}

		params := openai.ChatCompletionNewParams{
			Messages: openai.F(openAIMessages),
			Model:    openai.F(o.model),
			Tools: openai.F([]openai.ChatCompletionToolParam{
				{
					Type: openai.F(openai.ChatCompletionToolTypeFunction),
					Function: openai.F(openai.FunctionDefinitionParam{
						Name: openai.String("execute_tool"),
						Parameters: openai.F(openai.FunctionParameters{
							"type": "object",
							"properties": map[string]interface{}{
								"tool_name": map[string]string{
									"type": "string",
								},
								"arguments": map[string]string{
									"type": "string",
								},
							},
							"required": []string{"tool_name", "arguments"},
						}),
					}),
				},
			}),
		}

		stream := o.client.Chat.Completions.NewStreaming(ctx, params)
		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				// Handle tool calls
				if toolCalls := evt.Choices[0].Delta.ToolCalls; len(toolCalls) > 0 {
					for _, toolCall := range toolCalls {
						if toolCall.Function.Name != "" || toolCall.Function.Arguments != "" {
							events <- Event{
								Type: EventToolCall,
								Content: string(mustMarshal(map[string]interface{}{
									"name":      toolCall.Function.Name,
									"arguments": toolCall.Function.Arguments,
								})),
							}
						}
					}
					continue
				}

				// Handle regular content
				if content := evt.Choices[0].Delta.Content; content != "" {
					events <- Event{Type: EventToken, Content: content}
				}
			}
		}

		if err := stream.Err(); err != nil {
			events <- Event{Type: EventError, Error: err}
			return
		}

		events <- Event{Type: EventDone}
	}()

	return events
}

func (o *OpenAI) GenerateSync(ctx context.Context, messages []Message) (string, error) {
	openAIMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "user":
			openAIMessages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			openAIMessages[i] = openai.AssistantMessage(msg.Content)
		case "system":
			openAIMessages[i] = openai.SystemMessage(msg.Content)
		case "tool":
			openAIMessages[i] = openai.ToolMessage(msg.Name, msg.Content)
		}
	}

	params := openai.ChatCompletionNewParams{
		Messages: openai.F(openAIMessages),
		Model:    openai.F(o.model),
	}

	completion, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func (openai *OpenAI) Generate(ctx context.Context, messages []Message) <-chan Event {
	return openai.generateEvents(ctx, messages)
}
