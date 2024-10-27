package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	id           string
	role         types.Role
	state        types.AgentState
	tools        map[string]types.Tool
	buffer       *Buffer
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	messages     chan string
	context      string
	task         string
	toolset      *Toolset
	provider     provider.Provider
	capabilities map[string]func(context.Context, map[string]interface{}) (string, error)
	metrics      *AgentMetrics
}

type AgentMetrics struct {
	successRate   float64
	responseTime  time.Duration
	taskCount     int64
	lastOptimized time.Time
	mu            sync.RWMutex
}

// NewAgent creates a new agent with the given parameters
func NewAgent(id string, role types.Role, systemPrompt, userPrompt string, toolset *Toolset, provider provider.Provider) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		id:           id,
		role:         role,
		state:        types.StateIdle,
		tools:        toolset.GetToolsForRole(string(role)),
		buffer:       NewBuffer(systemPrompt, userPrompt),
		ctx:          ctx,
		cancel:       cancel,
		messages:     make(chan string, 100),
		context:      systemPrompt,
		task:         userPrompt,
		toolset:      toolset,
		provider:     provider,
		capabilities: make(map[string]func(context.Context, map[string]interface{}) (string, error)),
		metrics:      &AgentMetrics{},
	}
}

// GetID returns the agent's unique identifier
func (a *Agent) GetID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.id
}

// GetRole returns the agent's role
func (a *Agent) GetRole() types.Role {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.role
}

// GetState returns the agent's current state
func (a *Agent) GetState() types.AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// SetState updates the agent's state
func (a *Agent) SetState(state types.AgentState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state = state
}

// ReceiveMessage adds a message to the agent's message queue and buffer
func (a *Agent) ReceiveMessage(message string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Add to buffer first
	a.buffer.AddMessage("user", message)

	// Then try to add to channel
	select {
	case a.messages <- message:
		return nil
	default:
		return fmt.Errorf("message queue full for agent %s", a.id)
	}
}

// Shutdown gracefully shuts down the agent
func (a *Agent) Shutdown() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}

	close(a.messages)
	a.tools = nil
	a.state = types.StateDone
}

// GetContext returns the agent's context
func (a *Agent) GetContext() context.Context {
	return a.ctx
}

// GetTools returns the agent's available tools
func (a *Agent) GetTools() map[string]types.Tool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.tools
}

// ExecuteTask performs the agent's assigned task and returns the result
func (a *Agent) ExecuteTask() (string, error) {
	if a.provider == nil {
		return "", fmt.Errorf("provider not set for agent %s", a.id)
	}

	a.SetState(types.StateWorking)

	// Get messages from buffer
	bufferMsgs := a.buffer.GetMessages()
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
		return "", fmt.Errorf("no messages to process for agent %s", a.id)
	}

	response, err := a.provider.GenerateSync(a.ctx, messages)
	if err != nil {
		a.SetState(types.StateIdle)
		return "", fmt.Errorf("task execution failed: %w", err)
	}

	// Add the response to the buffer
	a.buffer.AddMessage("assistant", response)

	a.SetState(types.StateDone)
	return response, nil
}

// GetMessageCount returns the number of messages processed by this agent
func (a *Agent) GetMessageCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.messages)
}

// UpdateMetrics records performance metrics for the agent
func (a *Agent) UpdateMetrics(success bool, duration time.Duration) {
	a.metrics.mu.Lock()
	defer a.metrics.mu.Unlock()

	a.metrics.taskCount++
	a.metrics.responseTime = (a.metrics.responseTime + duration) / 2

	if success {
		// Weighted moving average for success rate
		a.metrics.successRate = (a.metrics.successRate * 0.8) + (1.0 * 0.2)
	} else {
		a.metrics.successRate = (a.metrics.successRate * 0.8) + (0.0 * 0.2)
	}
}

// GetPerformanceMetrics returns the current performance metrics
func (a *Agent) GetPerformanceMetrics() (float64, time.Duration, int64) {
	a.metrics.mu.RLock()
	defer a.metrics.mu.RUnlock()
	return a.metrics.successRate, a.metrics.responseTime, a.metrics.taskCount
}
