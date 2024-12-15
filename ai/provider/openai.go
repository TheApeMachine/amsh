package provider

import (
	"context"
	"fmt"
	"io"

	sdk "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

// Define a threshold for buffer processing
const bufferThreshold = 1024 * 4 // 4KB

type OpenAI struct {
	pr     *io.PipeReader
	pw     *io.PipeWriter
	client *sdk.Client
	model  string
}

/*
NewOpenAI creates an OpenAI provider, which is configurable with an endpoint,
api key, and model, given that a lot of other providers use the OpenAI API
specifications.
*/
func NewOpenAI(endpoint, apiKey, model string) *OpenAI {
	errnie.Trace("%s", "model", model)

	pr, pw := io.Pipe()
	return &OpenAI{
		pr: pr,
		pw: pw,
		client: sdk.NewClient(
			option.WithBaseURL(endpoint),
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

func (openai *OpenAI) Read(p []byte) (n int, err error) {
	errnie.Trace("OpenAI.Read START with buffer size: %d", len(p))

	if openai.pr == nil {
		errnie.Trace("OpenAI.Read pipe reader is nil")
		return 0, io.EOF
	}

	n, err = openai.pr.Read(p)
	errnie.Trace("OpenAI.Read got %d bytes, err: %v", n, err)

	if err == nil && n > 0 {
		// Try to unmarshal and log the content for debugging
		artifact := data.Empty()
		if err := artifact.Unmarshal(p[:n]); err == nil {
			errnie.Trace("OpenAI.Read content: %s", artifact.Peek("payload"))
		}
	}

	return n, err
}

func (openai *OpenAI) Write(p []byte) (n int, err error) {
	errnie.Trace("OpenAI.Write START with %d bytes", len(p))

	artifact := data.Empty()
	if err := artifact.Unmarshal(p); err != nil {
		errnie.Error(err)
		return 0, fmt.Errorf("failed to unmarshal artifact: %w", err)
	}
	errnie.Trace("OpenAI.Write unmarshaled artifact with payload: %s", artifact.Peek("payload"))

	// Process the request in a goroutine
	go func() {
		defer func() {
			errnie.Trace("OpenAI.Write goroutine CLEANUP")
			openai.pw.Close()
		}()

		errnie.Trace("OpenAI.Write preparing messages")
		messages := []sdk.ChatCompletionMessageParamUnion{
			openai.makeMessages(artifact),
		}

		errnie.Trace("OpenAI.Write starting stream")
		stream := openai.client.Chat.Completions.NewStreaming(context.Background(), sdk.ChatCompletionNewParams{
			Messages: sdk.F(messages),
			Model:    sdk.F(openai.model),
		})

		errnie.Trace("OpenAI.Write entering stream loop")
		for stream.Next() {
			if chunk := stream.Current(); len(chunk.Choices) > 0 {
				if delta := chunk.Choices[0].Delta; delta.Content != "" {
					errnie.Trace("OpenAI.Write received chunk: %s", delta.Content)

					responseArtifact := data.New(
						"openai", "assistant", "chunk", []byte(delta.Content),
					)

					errnie.Trace("OpenAI.Write marshaling response")
					buf := make([]byte, 1024)
					responseArtifact.Marshal(buf)

					errnie.Trace("OpenAI.Write writing chunk to pipe")
					if _, err := openai.pw.Write(buf); err != nil {
						errnie.Error(err)
						errnie.Trace("OpenAI.Write failed to write to pipe: %v", err)
						return
					}
					errnie.Trace("OpenAI.Write successfully wrote chunk")
				}
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
			errnie.Trace("OpenAI.Write stream error: %v", err)
			return
		}
		errnie.Trace("OpenAI.Write stream completed")
	}()

	errnie.Trace("OpenAI.Write returning with len: %d", len(p))
	return len(p), nil
}

func (openai *OpenAI) Close() error {
	errnie.Trace("%s", "OpenAI.Close", "close")

	if openai.pw != nil {
		openai.pw.Close()
	}

	if openai.pr != nil {
		openai.pr.Close()
	}

	openai.client = nil
	return nil
}

func (openai *OpenAI) makeMessages(artifact *data.Artifact) sdk.ChatCompletionMessageParamUnion {
	errnie.Trace("%s", "artifact", artifact.Peek("role"))

	switch artifact.Peek("role") {
	case "user":
		return sdk.UserMessage(artifact.Peek("payload"))
	case "assistant":
		return sdk.AssistantMessage(artifact.Peek("payload"))
	case "system":
		return sdk.SystemMessage(artifact.Peek("payload"))
	}

	return nil
}

func (openai *OpenAI) Generate(ctx context.Context, params GenerationParams) <-chan Event {
	errnie.Trace("%s", "OpenAI.Generate", "generate")

	events := make(chan Event, 64)

	go func() {
		defer close(events)

		openAIMessages := make([]sdk.ChatCompletionMessageParamUnion, len(params.Messages))
		for i, msg := range params.Messages {
			switch msg.Role {
			case "user":
				openAIMessages[i] = sdk.UserMessage(msg.Content)
			case "assistant":
				openAIMessages[i] = sdk.AssistantMessage(msg.Content)
			case "system":
				openAIMessages[i] = sdk.SystemMessage(msg.Content)
			case "tool":
				openAIMessages[i] = sdk.ToolMessage(msg.Name, msg.Content)
			}
		}

		stream := openai.client.Chat.Completions.NewStreaming(ctx, sdk.ChatCompletionNewParams{
			Messages: sdk.F(openAIMessages),
			Model:    sdk.F(openai.model),
			// Temperature:      openai.F(params.Temperature),
			// FrequencyPenalty: openai.F(params.FrequencyPenalty),
			// PresencePenalty:  openai.F(params.PresencePenalty),
		})

		for stream.Next() {
			evt := stream.Current()
			if len(evt.Choices) > 0 {
				events <- Event{Type: EventToken, Content: evt.Choices[0].Delta.Content}
			}
		}

		if err := stream.Err(); err != nil {
			errnie.Error(err)
			events <- Event{Type: EventError, Error: err}
			return
		}

		events <- Event{Type: EventDone}
	}()

	return events
}

func (openai *OpenAI) Configure(config map[string]interface{}) {
	// OpenAI-specific configuration can be added here if needed
}
