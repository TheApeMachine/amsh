package provider

import (
	"context"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
)

type Anthropic struct {
	client    *anthropic.Client
	model     string
	maxTokens int64
	system    string
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

func (a *Anthropic) Generate(artifacts []*data.Artifact) <-chan *data.Artifact {
	return twoface.NewAccumulator(
		"anthropic",
		"provider",
		"completion",
		artifacts...,
	).Yield(func(accumulator *twoface.Accumulator) {
		defer close(accumulator.Out)

		errnie.Log("===START===")
		requestParams := a.buildRequestParams(artifacts)
		stream := a.client.Messages.NewStreaming(context.Background(), requestParams)
		errnie.Log("===END===")

		for stream.Next() {
			event := stream.Current()

			switch event := event.AsUnion().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if event.Delta.Text != "" {
					response := data.New("anthropic", "assistant", a.model, []byte(event.Delta.Text))
					accumulator.Out <- response
				}
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
		}
	}).Generate()
}

func (a *Anthropic) buildRequestParams(artifacts []*data.Artifact) anthropic.MessageNewParams {
	messages := make([]anthropic.MessageParam, 0)
	var systemMessage string

	// First pass to extract system message and build regular messages
	for _, artifact := range artifacts {
		role := artifact.Peek("role")
		payload := artifact.Peek("payload")

		errnie.Log("Anthropic.Generate role %s payload %s", role, payload)

		if role == "system" {
			systemMessage = payload
			continue
		}

		var anthropicRole anthropic.MessageParamRole
		switch role {
		case "user":
			anthropicRole = anthropic.MessageParamRoleUser
		case "assistant":
			anthropicRole = anthropic.MessageParamRoleAssistant
		default:
			errnie.Warn("Anthropic.Generate unknown_role %s", role)
			continue
		}

		messages = append(messages, anthropic.MessageParam{
			Role: anthropic.F(anthropicRole),
			Content: anthropic.F([]anthropic.MessageParamContentUnion{
				anthropic.MessageParamContent{
					Type: anthropic.F(anthropic.MessageParamContentTypeText),
					Text: anthropic.F(payload),
				},
			}),
		})
	}

	// Build request params
	requestParams := anthropic.MessageNewParams{
		Model:       anthropic.F(a.model),
		Messages:    anthropic.F(messages),
		MaxTokens:   anthropic.F(a.maxTokens),
		Temperature: anthropic.F(0.7),
	}

	// Add system message if present (either from artifacts or Configure)
	if systemMessage != "" || a.system != "" {
		// Prefer system message from artifacts over configured one
		finalSystemMsg := systemMessage
		if finalSystemMsg == "" {
			finalSystemMsg = a.system
		}

		requestParams.System = anthropic.F([]anthropic.TextBlockParam{
			{
				Text: anthropic.F(finalSystemMsg),
				Type: anthropic.F(anthropic.TextBlockParamTypeText),
			},
		})
	}

	return requestParams
}

func (a *Anthropic) convertToAnthropicMessages(artifacts []*data.Artifact) []anthropic.MessageParam {
	anthropicMsgs := make([]anthropic.MessageParam, 0, len(artifacts))

	for _, artifact := range artifacts {
		role := artifact.Peek("role")
		payload := strings.TrimSpace(artifact.Peek("payload"))

		errnie.Log("Anthropic.Generate role %s payload %s", role, payload)

		var anthropicRole anthropic.MessageParamRole
		switch role {
		case "user":
			anthropicRole = anthropic.MessageParamRoleUser
		case "assistant":
			anthropicRole = anthropic.MessageParamRoleAssistant
		default:
			errnie.Warn("Anthropic.Generate unknown_role %s", role)
			continue
		}

		anthropicMsgs = append(anthropicMsgs, anthropic.MessageParam{
			Role: anthropic.F(anthropicRole),
			Content: anthropic.F([]anthropic.MessageParamContentUnion{
				anthropic.MessageParamContent{
					Type: anthropic.F(anthropic.MessageParamContentTypeText),
					Text: anthropic.F(payload),
				},
			}),
		})
	}

	return anthropicMsgs
}
