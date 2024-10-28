package ai

import (
	"context"
	"fmt"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/learning"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/reasoning"
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

	ctx        context.Context           `json:"-"`
	cancel     context.CancelFunc        `json:"-"`
	provider   provider.Provider         `json:"-"`
	mu         sync.RWMutex              `json:"-"`
	outputChan chan provider.Event       `json:"-"`
	reasoner   *reasoning.Engine         `json:"-"`
	learner    *learning.LearningAdapter `json:"-"`
}

// NewAgent creates a new agent with integrated reasoning and learning
func NewAgent(id string, role types.Role, systemPrompt, userPrompt string, tools map[string]types.Tool, prvdr provider.Provider) *Agent {
	log.Info("Creating new agent", "id", id, "role", role)
	ctx, cancel := context.WithCancel(context.Background())

	// Create reasoning engine for this agent
	validator := reasoning.NewValidator(reasoning.NewKnowledgeBase())
	metaReasoner := reasoning.NewMetaReasoner()
	reasoner := reasoning.NewEngine(validator, metaReasoner)

	return &Agent{
		ID:           id,
		Role:         role,
		State:        types.StateIdle,
		Context:      systemPrompt,
		Task:         userPrompt,
		Buffer:       NewBuffer(systemPrompt, userPrompt),
		Tools:        tools,
		Messages:     make(chan string, 100),
		outputChan:   make(chan provider.Event, 100),
		Capabilities: make(map[string]func(context.Context, map[string]interface{}) (string, error)),
		ctx:          ctx,
		cancel:       cancel,
		provider:     prvdr,
		reasoner:     reasoner,
		learner:      learning.NewLearningAdapter(),
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
	log.Info("Receiving message", "agent", a.ID)
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
	log.Info("Updating agent", "agent", a.ID)
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
	log.Info("Getting tools", "agent", a.ID)
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Tools
}

// ExecuteTaskStream now properly handles streaming responses
func (agent *Agent) ExecuteTaskStream() <-chan provider.Event {
	log.Info("Executing task stream", "agent", agent.ID)
	agent.mu.Lock()
	agent.State = types.StateWorking
	agent.mu.Unlock()

	go func() {
		defer close(agent.outputChan)
		defer func() {
			agent.mu.Lock()
			agent.State = types.StateDone
			agent.mu.Unlock()
		}()

		messages := agent.Buffer.GetMessages()
		for event := range agent.provider.Generate(agent.ctx, messages) {
			select {
			case <-agent.ctx.Done():
				return
			case agent.outputChan <- event:
				// Successfully sent event
			}
		}
	}()

	return agent.outputChan
}

// ExecuteTask now uses reasoning and learning
func (agent *Agent) ExecuteTask() (string, error) {
	log.Info("Executing task", "agent", agent.ID)
	if agent.provider == nil {
		return "", fmt.Errorf("provider not set for agent %s", agent.ID)
	}

	agent.SetState(types.StateWorking)

	// Get messages from buffer
	messages := agent.Buffer.GetMessages()

	// First, get the LLM's response
	var result string
	for event := range agent.provider.Generate(agent.ctx, messages) {
		if event.Error != nil {
			return "", event.Error
		}
		result += event.Content
	}

	// Optionally use reasoning to validate/enhance the response
	if agent.reasoner != nil {
		chain, err := agent.reasoner.ProcessReasoning(agent.ctx, result)
		if err == nil { // Only use reasoning result if successful
			result = agent.reasoner.FormatOutput(chain.Steps)
			// Record the outcome for learning
			agent.learner.RecordStrategyExecution(chain.Steps[len(chain.Steps)-1].Strategy, chain)
		}
	}

	// Add the response to the buffer
	agent.Buffer.AddMessage("assistant", result)

	agent.SetState(types.StateDone)
	return result, nil
}

// GetMessageCount returns the number of messages processed by this agent
func (a *Agent) GetMessageCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.Messages)
}

// HasCapability checks if the agent has a specific capability
func (agent *Agent) HasCapability(capability string) bool {
	log.Info("Checking agent capability", "agent", agent.ID, "capability", capability)
	agent.mu.RLock()
	defer agent.mu.RUnlock()

	// Check if the agent has this capability in its tools
	_, hasCapability := agent.Tools[capability]
	return hasCapability
}
