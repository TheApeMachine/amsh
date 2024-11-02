package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/theapemachine/amsh/errnie"
)

type Anthropic struct {
	client    *anthropic.Client
	model     string
	maxTokens int64
	system    string // Add system message field
}

func NewAnthropic(apiKey string, model string) *Anthropic {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
		option.WithHeader("x-api-key", apiKey),
	)
	return &Anthropic{
		client:    client,
		model:     model,
		maxTokens: 4096,
	}
}

func (a *Anthropic) Configure(config map[string]interface{}) {
	if systemMsg, ok := config["system_message"].(string); ok {
		a.system = systemMsg
	}
}

func (a *Anthropic) Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event {
	errnie.Info("generating with anthropic provider")
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		// Prepare the request parameters
		requestParams := anthropic.MessageNewParams{
			Model:       anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20240620),
			Messages:    anthropic.F(convertToAnthropicMessages(messages)),
			MaxTokens:   anthropic.F(a.maxTokens),
			Temperature: anthropic.F(params.Temperature),
		}

		// Only add system message if it's not empty
		if a.system != "" {
			requestParams.System = anthropic.F([]anthropic.TextBlockParam{{
				Text: anthropic.F(a.system),
				Type: anthropic.F(anthropic.TextBlockParamTypeText),
			}})
		}

		stream := a.client.Messages.NewStreaming(ctx, requestParams)
		message := anthropic.Message{}

		for stream.Next() {
			event := stream.Current()

			err := message.Accumulate(event)
			if err != nil {
				errnie.Error(err)
				events <- Event{Type: EventError, Error: err}
				return
			}

			switch event := event.AsUnion().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if event.Delta.Text != "" {
					events <- Event{Type: EventToken, Content: event.Delta.Text}
				}
			case anthropic.MessageStopEvent:
				events <- Event{Type: EventDone}
				return
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
			events <- Event{Type: EventError, Error: err}
			return
		}
	}()

	return events
}

func (a *Anthropic) GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error) {
	// Filter out system messages as they're handled separately
	var filteredMessages []Message
	for _, msg := range messages {
		if msg.Role != "system" {
			filteredMessages = append(filteredMessages, msg)
		}
	}

	// Prepare the request parameters
	requestParams := anthropic.MessageNewParams{
		Model:    anthropic.F(a.model),
		Messages: anthropic.F(convertToAnthropicMessages(filteredMessages)),
	}

	// Only add system message if it's not empty
	if a.system != "" {
		requestParams.System = anthropic.F([]anthropic.TextBlockParam{{
			Text: anthropic.F(a.system),
			Type: anthropic.F(anthropic.TextBlockParamTypeText),
		}})
	}

	message, err := a.client.Messages.New(ctx, requestParams)
	if err != nil {
		return "", err
	}

	return message.Content[0].Text, nil
}

// Helper function to convert our messages to Anthropic format
func convertToAnthropicMessages(msgs []Message) []anthropic.MessageParam {
	anthropicMsgs := make([]anthropic.MessageParam, len(msgs))
	for i, msg := range msgs {
		role := anthropic.MessageParamRoleUser
		if msg.Role == "assistant" {
			role = anthropic.MessageParamRoleAssistant
		}

		anthropicMsgs[i] = anthropic.MessageParam{
			Role: anthropic.F(role),
			Content: anthropic.F([]anthropic.MessageParamContentUnion{
				anthropic.MessageParamContent{
					Type: anthropic.F(anthropic.MessageParamContentTypeText),
					Text: anthropic.F(msg.Content),
				},
			}),
		}
	}

	return anthropicMsgs
}
