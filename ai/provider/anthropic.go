package provider

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/charmbracelet/log"
)

type Anthropic struct {
	client    *anthropic.Client
	model     string
	maxTokens int64
	system    string // Add system message field
}

func NewAnthropic(apiKey string, model string) *Anthropic {
	client := anthropic.NewClient()
	return &Anthropic{
		client:    client,
		model:     model,
		maxTokens: 2000,
	}
}

func (a *Anthropic) Configure(config map[string]interface{}) {
	if systemMsg, ok := config["system_message"].(string); ok {
		a.system = systemMsg
	}
}

func (a *Anthropic) Generate(ctx context.Context, messages []Message) <-chan Event {
	log.Info("generating with", "provider", "anthropic")
	events := make(chan Event, 64)

	go func() {
		defer close(events)

		// Parse the user's message from JSON
		var userMsg struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(messages[len(messages)-1].Content), &userMsg); err != nil {
			log.Error("Failed to parse user message", "error", err)
			events <- Event{Type: EventError, Error: err}
			return
		}

		// Create the messages array with the user's message
		processedMessages := []Message{
			{
				Role:    "user",
				Content: userMsg.Text,
			},
		}

		// Prepare the request parameters
		params := anthropic.MessageNewParams{
			Model:     anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20240620),
			Messages:  anthropic.F(convertToAnthropicMessages(processedMessages)),
			MaxTokens: anthropic.F(a.maxTokens),
		}

		// Only add system message if it's not empty
		if a.system != "" {
			params.System = anthropic.F([]anthropic.TextBlockParam{{
				Text: anthropic.F(a.system),
				Type: anthropic.F(anthropic.TextBlockParamTypeText),
			}})
		}

		stream := a.client.Messages.NewStreaming(ctx, params)
		message := anthropic.Message{}

		log.Info("Starting stream processing")

		for stream.Next() {
			event := stream.Current()

			err := message.Accumulate(event)
			if err != nil {
				log.Error("Error accumulating message", "error", err)
				events <- Event{Type: EventError, Error: err}
				return
			}

			switch event := event.AsUnion().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if event.Delta.Text != "" {
					events <- Event{Type: EventToken, Content: event.Delta.Text}
				}
			case anthropic.MessageStopEvent:
				log.Info("Stream completed")
				events <- Event{Type: EventDone}
				return
			}
		}

		if err := stream.Err(); err != nil {
			log.Error("Stream error", "error", err)
			events <- Event{Type: EventError, Error: err}
			return
		}

		log.Info("Generation completed")
	}()

	return events
}

func (a *Anthropic) GenerateSync(ctx context.Context, messages []Message) (string, error) {
	// Filter out system messages as they're handled separately
	var filteredMessages []Message
	for _, msg := range messages {
		if msg.Role != "system" {
			filteredMessages = append(filteredMessages, msg)
		}
	}

	// Prepare the request parameters
	params := anthropic.MessageNewParams{
		Model:    anthropic.F(a.model),
		Messages: anthropic.F(convertToAnthropicMessages(filteredMessages)),
	}

	// Only add system message if it's not empty
	if a.system != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{{
			Text: anthropic.F(a.system),
			Type: anthropic.F(anthropic.TextBlockParamTypeText),
		}})
	}

	message, err := a.client.Messages.New(ctx, params)
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

	log.Debug("Converted messages", "anthropicMsgs", anthropicMsgs)
	return anthropicMsgs
}
