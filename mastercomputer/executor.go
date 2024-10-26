package mastercomputer

import (
	"context"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
)

// Executor is responsible for executing tasks and interacting with tools.
// It now includes WorkerID and Events for better tracking and visualization.
type Executor struct {
	ctx      context.Context
	cancel   context.CancelFunc
	buffer   *Conversation
	toolset  *Toolset
	role     string
	workerID string
	events   *Events
	out      chan openai.ChatCompletionMessageParamUnion
}

// NewExecutor creates a new Executor instance with additional tracking parameters.
// It now accepts workerID and events channel.
func NewExecutor(buffer *Conversation, toolset *Toolset, role string, workerID string, out chan openai.ChatCompletionMessageParamUnion) *Executor {
	errnie.Trace()
	ctx, cancel := context.WithCancel(context.Background())

	return &Executor{
		ctx:      ctx,
		cancel:   cancel,
		buffer:   buffer,
		toolset:  toolset,
		role:     role,
		workerID: workerID,
		events:   NewEvents(),
		out:      out,
	}
}

// Start begins the execution process.
func (executor *Executor) Start() {
	errnie.Trace()
	if params, err := executor.prepareParams(executor.buffer.Truncate()); errnie.Error(err) == nil {
		completion := NewCompletion(executor.ctx)
		response := completion.Execute(params)

		// Add nil check before processing response
		if response == nil {
			log.Println("Received nil response from completion")
			executor.out <- openai.AssistantMessage("Error: Failed to get response from AI model")
			return
		}

		executor.processResponse(response)
	}
}

// prepareParams prepares the parameters for the OpenAI API call.
func (executor *Executor) prepareParams(messages []openai.ChatCompletionMessageParamUnion) (openai.ChatCompletionNewParams, error) {
	errnie.Trace()
	spew.Dump(messages)
	params := openai.ChatCompletionNewParams{
		Messages:    openai.F(messages),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(0.0),
		Tools:       openai.F(executor.toolset.Assign(executor.role)),
		Store:       openai.F(true),
	}

	return params, nil
}

// processResponse handles the response from the OpenAI API.
func (executor *Executor) processResponse(response *openai.ChatCompletion) {
	errnie.Trace()

	// Add defensive nil check at the start
	if response == nil {
		log.Println("Cannot process nil response")
		return
	}

	if len(response.Choices) == 0 {
		log.Println("No response from OpenAI")
		executor.out <- openai.AssistantMessage("Error: No response received from AI model")
		return
	}

	message := response.Choices[0].Message
	content := message.Content

	executor.events.channel <- Event{
		Timestamp: time.Now(),
		Type:      "ResponseReceived",
		Message:   content,
		WorkerID:  executor.workerID,
	}

	// Add response to worker's conversation buffer with worker-specific tagging
	if content != "" {
		executor.out <- openai.AssistantMessage(content)
	}

	executor.handleToolCalls(response)
}

// handleToolCalls processes any tool calls present in the response.
func (executor *Executor) handleToolCalls(response *openai.ChatCompletion) {
	errnie.Trace()
	executor.buffer.UpdateTokenCounts(response.Usage)
	message := response.Choices[0].Message

	if message.ToolCalls == nil || len(message.ToolCalls) == 0 {
		return
	}

	for _, toolCall := range message.ToolCalls {
		executor.out <- openai.AssistantMessage(message.Content)
		executor.out <- executor.toolset.Use(toolCall, executor.workerID)
	}
}
