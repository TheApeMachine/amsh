package mastercomputer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai/format"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/utils"
)

type Executor struct {
	ctx              context.Context
	cancel           context.CancelFunc
	worker           *Worker
	conversation     []openai.ChatCompletionMessageParamUnion
	maxContextTokens int
}

func NewExecutor(ctx context.Context, worker *Worker) *Executor {
	return &Executor{
		worker:           worker,
		maxContextTokens: 128000,
	}
}

func (e *Executor) Initialize() error {
	e.ctx, e.cancel = context.WithCancel(e.worker.ctx)
	return nil
}

func (e *Executor) Run() {
	defer e.Close()
	e.setupInitialConversation()

	params, err := e.prepareParams()
	if err != nil {
		log.Printf("Error preparing parameters: %v", err)
		return
	}

	response, err := e.executeCompletion(params)
	if err != nil {
		log.Printf("Error executing completion: %v", err)
		return
	}

	e.processResponse(response)
}

func (e *Executor) Close() {
	if e.cancel != nil {
		e.cancel()
	}
}

func (e *Executor) setupInitialConversation() {
	e.conversation = []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(e.worker.buffer.Peek("system")),
		openai.UserMessage(e.worker.buffer.Peek("user")),
	}
}

func (e *Executor) prepareParams() (openai.ChatCompletionNewParams, error) {
	messages := e.truncateConversation()

	memoryContent, err := e.retrieveFromShortTermMemory()
	if err != nil {
		log.Printf("Error retrieving from short-term memory: %v", err)
	} else if memoryContent != "" {
		// Include memory in the system prompt or as a message
		memoryMessage := openai.SystemMessage("Previous steps:\n" + memoryContent)
		messages = append([]openai.ChatCompletionMessageParamUnion{memoryMessage}, messages...)
	}

	responseFormat, err := e.getResponseFormat()
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	return openai.ChatCompletionNewParams{
		Messages:       openai.F(messages),
		ResponseFormat: openai.F(responseFormat),
		Tools:          openai.F(NewToolset(e.worker.buffer.Peek("workload")).tools),
		Model:          openai.F(openai.ChatModelGPT4oMini),
		Temperature:    openai.Float(0.0),
	}, nil
}

var semaphore = make(chan struct{}, 1)

func (e *Executor) executeCompletion(params openai.ChatCompletionNewParams) (*openai.ChatCompletion, error) {
	semaphore <- struct{}{}        // Acquire a token
	defer func() { <-semaphore }() // Release the token

	completion := NewCompletion(e.ctx)
	response, err := completion.Execute(e.ctx, params)
	if err != nil {
		var apiError *openai.Error
		if errors.As(err, &apiError) {
			switch apiError.StatusCode {
			case 429:
				log.Printf("Rate limit exceeded: %v", apiError)
				// Implement retry logic with exponential backoff
				time.Sleep(time.Minute) // Wait before retrying
				return nil, apiError
			case 401:
				log.Printf("Authentication error: %v", apiError)
				// Check API key and authentication
				return nil, apiError
			default:
				log.Printf("OpenAI API error: %v", apiError)
				return nil, apiError
			}
		} else {
			log.Printf("Unexpected error: %v", err)
			return nil, err
		}
	}

	return response, nil
}

func retryWithBackoff(operation func() error) error {
	backoff := time.Second
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		var apiError *openai.Error
		if errors.As(err, &apiError) && apiError.StatusCode == 429 {
			log.Printf("Rate limited. Retrying in %v...", backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return err
	}
	return errors.New("max retries exceeded")
}

func (e *Executor) processResponse(response *openai.ChatCompletion) {
	if response == nil || len(response.Choices) == 0 {
		log.Println("No response from OpenAI")
		return
	}

	message := response.Choices[0].Message
	content := message.Content
	if content == "" {
		log.Println("Empty response content")
		return
	}

	if err := e.printResponse(content); err != nil {
		log.Printf("Error printing response: %v", err)
	}

	e.conversation = append(e.conversation, openai.AssistantMessage(content))
	e.storeInShortTermMemory(content)
	e.storeInLongTermMemory(content)

	// Handle tool calls if any
	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			e.handleToolCall(toolCall)
		}
	}
}

func (e *Executor) storeInShortTermMemory(content string) {
	artifact := data.New(e.worker.ID, "memory", "short-term", []byte(content))
	e.worker.memory.Write(artifact.Marshal())
}

func (e *Executor) storeInLongTermMemory(content string) {
	// Create an artifact for long-term memory
	artifact := data.New(e.worker.ID, "memory", "long-term", []byte(content))
	// For this example, we'll store in both vector and graph stores
	artifact.Poke("scope", "vector") // Or "graph" for Neo4j, or both

	// Write to long-term memory
	e.worker.memory.Write(artifact.Marshal())
}

func (e *Executor) retrieveFromShortTermMemory() (string, error) {
	buffer := make([]byte, e.worker.memory.ShortTerm.Length())
	n, err := e.worker.memory.ShortTerm.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	return string(buffer[:n]), nil
}

func (e *Executor) printResponse(content string) error {
	workload := e.worker.buffer.Peek("workload")
	var respFormat format.ResponseFormat

	switch workload {
	case "reasoning":
		var strategy format.Strategy
		if err := json.Unmarshal([]byte(content), &strategy); err != nil {
			return err
		}
		respFormat = strategy
	case "messaging":
		var msg format.Messaging
		if err := json.Unmarshal([]byte(content), &msg); err != nil {
			return err
		}
		respFormat = msg
	default:
		return errors.New("unknown workload")
	}

	fmt.Println(respFormat.String())
	return nil
}

func (e *Executor) handleToolCall(toolCall openai.ChatCompletionMessageToolCall) {
	switch toolCall.Function.Name {
	case "publish_message":
		// Handle publishing a message
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Printf("Error unmarshalling tool arguments: %v", err)
			return
		}
		e.publishMessage(args)
	case "worker":
		// Handle creating a new worker
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			log.Printf("Error unmarshalling tool arguments: %v", err)
			return
		}
		e.createWorker(args)
	default:
		log.Printf("Unknown tool: %s", toolCall.Function.Name)
	}
}

func (e *Executor) publishMessage(args map[string]interface{}) {
	topic, _ := args["topic"].(string)
	messageContent, _ := args["message"].(string)
	message := data.New(e.worker.ID, "message", "publish", []byte(messageContent))
	message.Poke("origin", e.worker.ID)
	message.Poke("scope", topic)
	e.worker.queue.Publish(message)
}

func (e *Executor) createWorker(args map[string]interface{}) {
	systemPrompt, _ := args["system"].(string)
	userPrompt, _ := args["user"].(string)
	formatStr, _ := args["format"].(string)
	ID := utils.NewID()

	artifact := data.New(utils.NewID(), "buffer", "setup", nil)
	artifact.Poke("id", ID)
	artifact.Poke("system", systemPrompt)
	artifact.Poke("user", userPrompt)
	artifact.Poke("workload", formatStr)
	artifact.Poke("payload", fmt.Sprintf("Agent %s is ready to go!", ID))

	newWorker := NewWorker(e.worker.parentCtx, artifact, e.worker.manager)
	go func() {
		if err := newWorker.Initialize(); err != nil {
			log.Printf("Worker %s initialization failed: %v", ID, err)
		}
	}()
}

func (e *Executor) getResponseFormat() (openai.ChatCompletionNewParamsResponseFormatUnion, error) {
	workload := e.worker.buffer.Peek("workload")
	switch workload {
	case "reasoning":
		return openai.ResponseFormatJSONSchemaParam{
			Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
			JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        openai.F("reasoning"),
				Description: openai.F("Available reasoning strategies"),
				Schema:      openai.F(GenerateSchema[format.Strategy]()),
			}),
		}, nil
	case "messaging":
		return openai.ResponseFormatJSONSchemaParam{
			Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
			JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        openai.F("messaging"),
				Description: openai.F("Messaging format"),
				Schema:      openai.F(GenerateSchema[format.Messaging]()),
			}),
		}, nil
	default:
		return nil, errors.New("unsupported workload")
	}
}

func (e *Executor) truncateConversation() []openai.ChatCompletionMessageParamUnion {
	maxTokens := e.maxContextTokens - 500
	totalTokens := 0
	var truncatedMessages []openai.ChatCompletionMessageParamUnion

	for i := len(e.conversation) - 1; i >= 0; i-- {
		msg := e.conversation[i]
		tokens := e.estimateTokens(msg)
		if totalTokens+tokens > maxTokens {
			break
		}
		truncatedMessages = append([]openai.ChatCompletionMessageParamUnion{msg}, truncatedMessages...)
		totalTokens += tokens
	}

	return truncatedMessages
}

func (e *Executor) estimateTokens(msg openai.ChatCompletionMessageParamUnion) int {
	content, role := extractContentAndRole(msg)
	// Implement token estimation logic
	return len(content) + len(role) // Simplified estimation
}

func extractContentAndRole(msg openai.ChatCompletionMessageParamUnion) (content, role string) {
	switch m := msg.(type) {
	case openai.ChatCompletionUserMessageParam:
		return m.Content.String(), "user"
	case openai.ChatCompletionAssistantMessageParam:
		return m.Content.String(), "assistant"
	case openai.ChatCompletionSystemMessageParam:
		return m.Content.String(), "system"
	default:
		return "", ""
	}
}
