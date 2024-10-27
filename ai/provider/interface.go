package provider

import (
	"context"
)

// Event represents different types of provider events
type Event struct {
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

// Provider defines the interface for AI providers
type Provider interface {
	// Generate returns a channel of events (tokens, tool calls, errors)
	Generate(ctx context.Context, messages []Message) <-chan Event

	// GenerateSync generates a complete response synchronously
	GenerateSync(ctx context.Context, messages []Message) (string, error)

	// Configure allows provider-specific configuration
	Configure(config map[string]interface{})
}

// Message represents a chat message
type Message struct {
	Role     string                 `json:"role"`
	Content  string                 `json:"content"`
	Name     string                 `json:"name,omitempty"`
	Function map[string]interface{} `json:"function,omitempty"`
}
