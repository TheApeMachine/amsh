package types

import "context"

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

// Role represents an agent's role in the system
type Role string

const (
	RoleResearcher Role = "researcher"
	RoleAnalyst    Role = "analyst"
)

// Tool represents a function that can be called by an AI agent
type Tool interface {
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
	GetSchema() ToolSchema
}

// ToolSchema defines the structure and requirements of a tool
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolHandler is a function that implements the actual tool logic
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// AgentContext provides access to agent-specific functionality
type AgentContext interface {
	SetState(state AgentState)
	GetState() AgentState
	GetRole() Role
	ReceiveMessage(message string) error // Add this line
}

// TeamManager interface defines the methods required for team management
type TeamManager interface {
	GetNextAgentID() int
	AddAgent(AgentContext) error
	GetToolset() ToolsetManager
	GetTeam(name string) (*Team, error) // Updated to match implementation
}

// ToolsetManager defines the methods required for toolset management
type ToolsetManager interface {
	GetToolsForRole(role string) map[string]Tool
	RegisterToolHandler(name string, handler ToolHandler) error
}

// LogicalExpression represents a logical statement or operation
type LogicalExpression struct {
	Operation     LogicalOperation
	Operands      []interface{}
	Confidence    float64
	Verifications []VerificationStep
	Content       string // Add this field to store the textual content
}

// Team represents a group of agents working together
type Team struct {
	Name   string
	Agents map[string]AgentContext
}
