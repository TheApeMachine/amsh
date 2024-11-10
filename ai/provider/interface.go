package provider

import (
	"context"
)

// Event represents different types of provider events
type Event struct {
	TeamID  string
	AgentID string
	Type    EventType
	Content string
	Error   error
}

type EventType int

const (
	EventToken EventType = iota
	EventToolCall
	EventFunctionCall
	EventError
	EventDone
)

type Message struct {
	Role     string                 `json:"role"`
	Content  string                 `json:"content"`
	Name     string                 `json:"name,omitempty"`
	Function map[string]interface{} `json:"function,omitempty"`
}

type GenerationParams struct {
	Messages               []Message
	Temperature            float64
	TopP                   float64
	TopK                   int
	Interestingness        float64
	InterestingnessHistory []float64
}

// Provider defines the interface for AI providers
type Provider interface {
	// Generate returns a channel of events (tokens, tool calls, errors)
	Generate(ctx context.Context, params GenerationParams) <-chan Event

	// Configure allows provider-specific configuration
	Configure(config map[string]interface{})
}
