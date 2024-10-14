package mastercomputer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/charmbracelet/log"
	"github.com/openai/openai-go"
	"github.com/pkoukk/tiktoken-go"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type Executor struct {
	pctx             context.Context
	ctx              context.Context
	cancel           context.CancelFunc
	worker           *Worker
	context          []openai.ChatCompletionMessageParamUnion
	maxContextTokens int
	tokenCounts      []int64 // Store token counts for each message
	done             bool
	pr               *io.PipeReader
	pw               *io.PipeWriter
}

func (executor *Executor) Initialize() *Executor {
	executor.ctx, executor.cancel = context.WithCancel(executor.pctx)
	executor.context = []openai.ChatCompletionMessageParamUnion{}
	executor.tokenCounts = []int64{}
	executor.done = false
	executor.pr, executor.pw = io.Pipe()

	return executor
}

func NewExecutor(ctx context.Context, worker *Worker) *Executor {
	return &Executor{
		pctx:             ctx,
		worker:           worker,
		context:          []openai.ChatCompletionMessageParamUnion{},
		maxContextTokens: 128000,
		tokenCounts:      []int64{},
		done:             false,
	}
}

func (executor *Executor) Read(p []byte) (n int, err error) {
	if executor.pr == nil || executor.pw == nil {
		return 0, io.ErrNoProgress
	}

	return executor.pr.Read(p)
}

func (executor *Executor) Write(p []byte) (n int, err error) {
	if executor.pr == nil || executor.pw == nil {
		return 0, io.ErrNoProgress
	}

	artifact := data.Empty
	artifact.Write(p)

	executor.context = []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(artifact.Peek("system")),
		openai.UserMessage(artifact.Peek("user")),
	}

	for {
		select {
		case <-executor.pctx.Done():
			return
		case <-executor.ctx.Done():
			return
		default:
			if executor.done {
				return
			}

			params := executor.getParamsWithManagedContext()

			response, err := executor.executeCompletion(params)
			if err != nil {
				log.Error("Error executing completion", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			executor.processResponse(response)
			time.Sleep(1 * time.Second)
		}
	}
}

func (executor *Executor) Close() error {
	return executor.pw.Close()
}

func (executor *Executor) executeCompletion(params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	completion := NewCompletion(executor.ctx)
	response, err := completion.Execute(executor.ctx, params)

	if err != nil {
		return nil, errnie.Error(err)
	}

	if response == nil {
		return nil, errnie.Error(errors.New("received nil response from OpenAI"))
	}

	return response, nil
}

func (executor *Executor) processResponse(response *openai.ChatCompletion) {
	if response.Usage.CompletionTokens > 0 {
		executor.updateTokenCounts(response.Usage)
	}

	userMessage, err := executor.extractAndPrintResponse(response)
	if err != nil {
		log.Error(err)
		return
	}

	if userMessage != nil {
		executor.updateConversationLog(userMessage)
	}

	toolMessage := executor.handleToolCalls(response)
	if toolMessage != nil {
		executor.updateConversationLog(toolMessage)
	}
}

var formats = map[string]format.ResponseFormat{
	"reasoning":   format.Strategy{},
	"environment": format.Environment{},
	"messaging":   format.Messaging{},
}

var formatMap = map[string]openai.ResponseFormatJSONSchemaJSONSchemaParam{
	"reasoning": {
		Name:        openai.F("reasoning"),
		Description: openai.F("Available reasoning strategies"),
		Schema:      openai.F(GenerateSchema[format.Strategy]()),
		Strict:      openai.Bool(false),
	},
	"environment": {
		Name:        openai.F("environment"),
		Description: openai.F("Available environment commands"),
		Schema:      openai.F(GenerateSchema[format.Environment]()),
		Strict:      openai.Bool(true),
	},
	"messaging": {
		Name:        openai.F("messaging"),
		Description: openai.F("Available messaging commands"),
		Schema:      openai.F(GenerateSchema[format.Messaging]()),
		Strict:      openai.Bool(true),
	},
}

func (executor *Executor) extractAndPrintResponse(response *openai.ChatCompletion) (openai.ChatCompletionMessageParamUnion, error) {
	if response == nil || len(response.Choices) == 0 {
		log.Error("No response from OpenAI")
		return nil, nil
	}

	content := response.Choices[0].Message.Content
	if content == "" {
		return nil, nil
	}

	format := formats[executor.worker.buffer.Peek("workload")].Format()
	if err := json.Unmarshal([]byte(content), &format); err != nil {
		return nil, err
	}

	return openai.AssistantMessage(content), nil
}

func (executor *Executor) handleToolCalls(response *openai.ChatCompletion) openai.ChatCompletionMessageParamUnion {
	if response == nil || len(response.Choices) == 0 {
		log.Error("No response from OpenAI")
		return nil
	}

	message := response.Choices[0].Message
	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		log.Error("No tool calls in response")
		return nil
	}

	for _, toolCall := range message.ToolCalls {
		var args map[string]any
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Error("Error unmarshalling tool call arguments", "error", err)
			continue
		}

		switch toolCall.Function.Name {
		case "":
		}
	}

	log.Warn("No tool calls matched")
	return nil
}

func (executor *Executor) getParamsWithManagedContext() openai.ChatCompletionNewParams {
	// Truncate conversation to fit within token limit
	messages := executor.truncateConversation()

	return openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(
					formatMap[executor.worker.buffer.Peek("workload")],
				),
			},
		),
		Tools:       openai.F(NewToolset("reasoning").tools),
		Model:       openai.F(openai.ChatModelGPT4o),
		Temperature: openai.Float(0.0),
	}
}

func (executor *Executor) truncateConversation() []openai.ChatCompletionMessageParamUnion {
	maxTokens := executor.maxContextTokens - 500 // Reserve tokens for response
	totalTokens := 0
	var truncatedMessages []openai.ChatCompletionMessageParamUnion

	// Start from the most recent message
	for i := len(executor.context) - 1; i >= 0; i-- {
		msg := executor.context[i]
		messageTokens := executor.estimateTokens(msg)
		if totalTokens+messageTokens <= maxTokens {
			truncatedMessages = append([]openai.ChatCompletionMessageParamUnion{msg}, truncatedMessages...)
			totalTokens += messageTokens
		} else {
			break
		}
	}

	return truncatedMessages
}

func (executor *Executor) updateTokenCounts(usage openai.CompletionUsage) {
	executor.tokenCounts = append(executor.tokenCounts, usage.TotalTokens)
}

func (executor *Executor) updateConversationLog(message openai.ChatCompletionMessageParamUnion) {
	executor.context = append(executor.context, message)
}

func (executor *Executor) estimateTokens(msg openai.ChatCompletionMessageParamUnion) int {
	content := ""
	role := ""
	switch m := msg.(type) {
	case openai.ChatCompletionUserMessageParam:
		content = m.Content.String()
		role = "user"
	case openai.ChatCompletionAssistantMessageParam:
		content = m.Content.String()
		role = "assistant"
	case openai.ChatCompletionSystemMessageParam:
		content = m.Content.String()
		role = "system"
	case openai.ChatCompletionToolMessageParam:
		content = m.Content.String()
		role = "function"
	}

	// Use tiktoken-go to estimate tokens
	encoding, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		log.Error("Error getting encoding", "error", err)
		return 0
	}

	tokensPerMessage := 4 // As per OpenAI's token estimation guidelines

	numTokens := tokensPerMessage
	numTokens += len(encoding.Encode(content, nil, nil))
	if role == "user" || role == "assistant" || role == "system" || role == "function" {
		numTokens += len(encoding.Encode(role, nil, nil))
	}

	return numTokens
}
