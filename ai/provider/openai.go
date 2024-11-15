package provider

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
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
		maxTokens: 4096,
	}
}

func (o *OpenAI) Generate(ctx context.Context, params GenerationParams) <-chan Event {
	errnie.Info("generating with %s", o.model)
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		openAIMessages := make([]openai.ChatCompletionMessageParamUnion, len(params.Messages))
		for i, msg := range params.Messages {
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

		stream := o.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages:         openai.F(openAIMessages),
			Model:            openai.F(o.model),
			Temperature:      openai.F(params.Temperature),
			FrequencyPenalty: openai.F(params.FrequencyPenalty),
			PresencePenalty:  openai.F(params.PresencePenalty),
		})

		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				events <- Event{Type: EventToken, Content: evt.Choices[0].Delta.Content}
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

// Add Configure method
func (o *OpenAI) Configure(config map[string]interface{}) {
	// OpenAI-specific configuration can be added here if needed
}
