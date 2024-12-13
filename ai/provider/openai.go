package provider

import (
	"bytes"
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
	buffer *bytes.Buffer
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
		pr:     pr,
		pw:     pw,
		buffer: bytes.NewBuffer(nil),
		client: sdk.NewClient(
			option.WithBaseURL(endpoint),
			option.WithAPIKey(apiKey),
		),
		model: model,
	}
}

func (openai *OpenAI) Read(p []byte) (n int, err error) {
	artifact := data.Empty()
	artifact.Unmarshal(p)
	errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))

	if openai.pr == nil {
		return 0, io.EOF
	}

	// Read into buffer first
	if n = errnie.SafeMust(func() (int, error) {
		return openai.pr.Read(p)
	}); n == 0 {
		return 0, io.EOF
	}

	openai.buffer.Write(p[:n])

	// Process when we have enough data or on EOF
	if err == io.EOF || openai.buffer.Len() > bufferThreshold {
		openai.processBuffer(context.Background())
	}

	return n, nil
}

func (openai *OpenAI) processBuffer(ctx context.Context) {
	errnie.Trace("%s", "OpenAI.processBuffer", "processing buffer")

	buf := openai.buffer.Bytes()
	openai.buffer.Reset()

	// Convert bytes back to Artifact using package function
	artifact := data.Empty()
	artifact.Unmarshal(buf)

	errnie.Trace("%s", "OpenAI.processBuffer", fmt.Sprintf("processing message from role: %s", artifact.Peek("role")))

	messages := []sdk.ChatCompletionMessageParamUnion{
		openai.makeMessages(artifact),
	}

	stream := openai.client.Chat.Completions.NewStreaming(ctx, sdk.ChatCompletionNewParams{
		Messages: sdk.F(messages),
		Model:    sdk.F(openai.model),
	})

	var (
		chunk sdk.ChatCompletionChunk
		delta sdk.ChatCompletionChunkChoicesDelta
	)

	for stream.Next() {
		if chunk = stream.Current(); len(chunk.Choices) > 0 {
			if delta = chunk.Choices[0].Delta; delta.Content != "" {
				errnie.Trace("%s", "OpenAI.processBuffer", fmt.Sprintf("received chunk: %d bytes", len(delta.Content)))
				_ = errnie.SafeMust(func() (int64, error) {
					return io.Copy(openai.pw, data.New(
						"openai", "assistant", "chunk", []byte(delta.Content),
					))
				})
			}
		}
	}

	if err := stream.Err(); err != nil {
		errnie.Error(err)
		return
	}

	errnie.Trace("%s", "OpenAI.processBuffer", "completed processing buffer")
}

func (openai *OpenAI) Write(p []byte) (n int, err error) {
	artifact := data.Empty()
	artifact.Unmarshal(p)
	errnie.Trace("%s", "artifact.Payload", artifact.Peek("payload"))

	if openai.pw == nil {
		return 0, io.ErrClosedPipe
	}

	// Write directly to pipe
	n, err = openai.pw.Write(p)
	if err != nil {
		if err == io.EOF {
			openai.pw = nil
		}
		errnie.Trace("%s", "OpenAI.Write", "completed with status: "+err.Error())
		return n, err
	}

	errnie.Trace("%s", "OpenAI.Write", fmt.Sprintf("wrote %d bytes", n))
	return n, nil
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
