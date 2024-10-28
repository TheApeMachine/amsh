package types

// Agent defines the interface for agent behavior
type Agent interface {
	GetID() string
	GetRole() Role
	GetState() AgentState
	SetState(state AgentState)
	ExecuteTask() (string, error)
	ExecuteTaskStream() <-chan Event
	GetTools() map[string]Tool
	ReceiveMessage(message string) error
	Shutdown()
}

// Event represents a provider event
type Event interface {
	GetContent() string
	GetError() error
	IsDone() bool
}

// AgentManager handles agent lifecycle and management
type AgentManager interface {
	CreateAgent(config AgentConfig) (Agent, error)
	GetAgent(id string) (Agent, error)
	DeleteAgent(id string) error
}

// AgentConfig defines the configuration for creating a new agent
type AgentConfig struct {
	Role           string          `json:"role"`
	Specialization string          `json:"specialization"`
	SystemPrompt   string          `json:"system_prompt"`
	UserPrompt     string          `json:"user_prompt"`
	Tools          map[string]Tool `json:"tools,omitempty"`
}
