package provider

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
)

type Anthropic struct {
	client    *anthropic.Client
	model     string
	maxTokens int
}

func NewAnthropic(apiKey string, model string) *Anthropic {
	client := anthropic.NewClient()
	return &Anthropic{
		client:    client,
		model:     model,
		maxTokens: 2000,
	}
}

func (a *Anthropic) Generate(ctx context.Context, messages []Message) <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)

		stream := a.client.Messages.NewStreaming(context.TODO(), anthropic.MessageNewParams{
			Model:    anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20240620),
			Messages: anthropic.F(convertToAnthropicMessages(messages)),
		})

		message := anthropic.Message{}
		for stream.Next() {
			event := stream.Current()
			err := message.Accumulate(event)
			if err != nil {
				panic(err)
			}

			switch event := event.AsUnion().(type) {
			case anthropic.ContentBlockStartEvent:
				if event.ContentBlock.Name != "" {
					print(event.ContentBlock.Name + ": ")
				}
			case anthropic.ContentBlockDeltaEvent:
				print(event.Delta.Text)
				print(event.Delta.PartialJSON)
			case anthropic.ContentBlockStopEvent:
				println()
				println()
			case anthropic.MessageStopEvent:
				println()
			}
		}

		if stream.Err() != nil {
			panic(stream.Err())
		}
	}()

	return events
}

func (a *Anthropic) GenerateSync(ctx context.Context, messages []Message) (string, error) {
	message, err := a.client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		Model:    anthropic.F(a.model),
		Messages: anthropic.F(convertToAnthropicMessages(messages)),
	})

	if err != nil {
		return "", err
	}

	return message.Content[0].Text, nil
}

// Add this helper function
func convertToAnthropicMessages(msgs []Message) []anthropic.MessageParam {
	anthropicMsgs := make([]anthropic.MessageParam, len(msgs))
	for i, msg := range msgs {
		anthropicMsgs[i] = anthropic.MessageParam{
			Role: anthropic.F(anthropic.MessageParamRole(msg.Role)), // Use proper Role type
			Content: anthropic.F([]anthropic.MessageParamContentUnion{ // Use proper Content type
				anthropic.MessageParamContent{
					Text: anthropic.F(msg.Content),
					Type: anthropic.F(anthropic.MessageParamContentTypeText),
				},
			}),
		}
	}
	return anthropicMsgs
}
