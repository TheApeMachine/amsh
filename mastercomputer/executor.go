package mastercomputer

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/utils"
)

// Executor is responsible for executing completions and managing scoped context.
type Executor struct {
	ctx          context.Context
	cancel       context.CancelFunc
	sequencer    *Sequencer
	conversation *Conversation
	toolset      *Toolset
}

func NewExecutor(sequencer *Sequencer) *Executor {
	errnie.Trace()
	ctx, cancel := context.WithCancel(context.Background())

	return &Executor{
		ctx:          ctx,
		cancel:       cancel,
		conversation: NewConversation(),
		sequencer:    sequencer,
		toolset:      NewToolset(),
	}
}

func (executor *Executor) Do(worker *Worker) {
	errnie.Trace()

	// Preserve existing conversation context, scoped by the worker
	messages := worker.buffer.GetScopedMessages()

	if len(messages) == 0 {
		// Add initial system and user prompts to scoped context
		worker.buffer.AddMessage(openai.SystemMessage(worker.system))
		worker.buffer.AddMessage(openai.UserMessage(worker.user))
	}

	// Prepare parameters for API request
	if params, err := executor.prepareParams(worker); errnie.Error(err) == nil {
		response, err := executor.executeCompletion(params)
		if errnie.Error(err) == nil {
			executor.processResponse(worker, response)
		}
	}
}

var semaphore = make(chan struct{}, 1)

func (executor *Executor) executeCompletion(params openai.ChatCompletionNewParams) (response *openai.ChatCompletion, err error) {
	errnie.Trace()

	semaphore <- struct{}{}
	defer func() { <-semaphore }()
	completion := NewCompletion(executor.ctx)
	return completion.Execute(executor.ctx, params)
}

func (executor *Executor) prepareParams(worker *Worker) (openai.ChatCompletionNewParams, error) {
	errnie.Trace()

	messages := worker.buffer.GetScopedMessages()
	tools := worker.toolset

	params := openai.ChatCompletionNewParams{
		Messages:    openai.F(messages),
		Model:       openai.F(openai.ChatModelGPT4oMini),
		Temperature: openai.Float(utils.ToFixed(worker.temperature, 1)),
		Tools:       openai.F(tools),
		Store:       openai.F(true),
	}

	return params, nil
}

func (executor *Executor) processResponse(worker *Worker, response *openai.ChatCompletion) {
	errnie.Trace()

	if len(response.Choices) == 0 {
		log.Println("No response from OpenAI")
		return
	}

	message := response.Choices[0].Message
	content := message.Content

	// Add response to worker's conversation buffer with worker-specific tagging
	if content != "" {
		taggedMessage := openai.AssistantMessage(fmt.Sprintf("[From %s]: %s", worker.name, content))
		worker.buffer.AddMessage(taggedMessage)
		executor.sequencer.output.Console(worker, MsgTypeAssistant, content)
	}
}
