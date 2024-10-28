package ai

import (
	"context"
	"fmt"
	"sync"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
)

// AgentState represents the current state of an agent
type AgentState string

const (
	StateIdle      AgentState = "idle"
	StateThinking  AgentState = "thinking"
	StateWorking   AgentState = "working"
	StateWaiting   AgentState = "waiting"
	StateReviewing AgentState = "reviewing"
	StateDone      AgentState = "done"
)

// Agent represents an AI agent that can perform tasks and communicate with other agents
type Agent struct {
	ID           string                                                                   `json:"id"`
	Role         types.Role                                                               `json:"role"`
	State        types.AgentState                                                         `json:"state"`
	Context      string                                                                   `json:"context"`
	Task         string                                                                   `json:"task"`
	Buffer       *Buffer                                                                  `json:"buffer"`
	Tools        map[string]types.Tool                                                    `json:"tools"`
	Messages     chan string                                                              `json:"-"`
	Capabilities map[string]func(context.Context, map[string]interface{}) (string, error) `json:"-"`

	ctx      context.Context    `json:"-"`
	cancel   context.CancelFunc `json:"-"`
	provider provider.Provider  `json:"-"`
	mu       sync.RWMutex       `json:"-"`
}

// NewAgent creates a new agent with the given parameters
func NewAgent(id string, role types.Role, systemPrompt, userPrompt string, tools map[string]types.Tool, provider provider.Provider) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		ID:           id,
		Role:         role,
		State:        types.StateIdle,
		Context:      systemPrompt,
		Task:         userPrompt,
		Buffer:       NewBuffer(systemPrompt, userPrompt),
		Tools:        tools,
		Messages:     make(chan string, 100),
		Capabilities: make(map[string]func(context.Context, map[string]interface{}) (string, error)),
		ctx:          ctx,
		cancel:       cancel,
		provider:     provider,
	}
}

// GetID returns the agent's unique identifier
func (a *Agent) GetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ID
}

// GetRole returns the agent's role
func (a *Agent) GetRole() types.Role {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Role
}

// GetState returns the agent's current state
func (a *Agent) GetState() types.AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.State
}

// SetState updates the agent's state
func (a *Agent) SetState(state types.AgentState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State = state
}

func (a *Agent) GetBuffer() *Buffer {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Buffer
}

// ReceiveMessage adds a message to the agent's message queue and buffer
func (a *Agent) ReceiveMessage(message string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Add to buffer first
	a.Buffer.AddMessage("user", message)

	// Then try to add to channel
	select {
	case a.Messages <- message:
		return nil
	default:
		return fmt.Errorf("message queue full for agent %s", a.ID)
	}
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}

	close(a.Messages)
	a.Tools = nil
	a.State = types.StateDone
}

func (a *Agent) Update(userMessages string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Buffer.AddMessage("user", userMessages)
}

// GetContext returns the agent's context
func (a *Agent) GetContext() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Context
}

// GetTools returns the agent's available tools
func (a *Agent) GetTools() map[string]types.Tool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Tools
}

func (a *Agent) ExecuteTaskStream() <-chan provider.Event {
	responseChan := make(chan provider.Event)

	go func() {
		defer close(responseChan)

		for event := range a.provider.Generate(a.ctx, a.Buffer.GetMessages()) {
			fmt.Print(event)
			responseChan <- event
		}
	}()

	return responseChan
}

// ExecuteTask performs the agent's assigned task and returns the result
func (a *Agent) ExecuteTask() (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not set for agent %s", a.ID)
	}

	a.SetState(types.StateWorking)

	// Get messages from buffer
	bufferMsgs := a.Buffer.GetMessages()
	messages := make([]provider.Message, 0, len(bufferMsgs))

	// Convert Buffer messages to Provider messages
	for _, msg := range bufferMsgs {
		messages = append(messages, provider.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Ensure we have messages to process
	if len(messages) == 0 {
		a.SetState(types.StateIdle)
		return "", fmt.Errorf("no messages to process for agent %s", a.ID)
	}

	// Use the streaming interface and collect the results
	var result string
	for event := range a.provider.Generate(a.ctx, messages) {
		if event.Error != nil {
			a.SetState(types.StateIdle)
			return "", fmt.Errorf("task execution failed: %w", event.Error)
		}
		result += event.Content
	}

	// Add the response to the buffer
	a.Buffer.AddMessage("assistant", result)

	a.SetState(types.StateDone)
	return result, nil
}

// GetMessageCount returns the number of messages processed by this agent
func (a *Agent) GetMessageCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.Messages)
}
