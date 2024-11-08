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

type GenerationParams struct {
	Temperature      float64 `json:"temperature" jsonschema:"title=Temperature,description=The temperature of the generation,required"`
	FrequencyPenalty float64 `json:"frequency_penalty" jsonschema:"title=Frequency Penalty,description=The frequency penalty of the generation,required"`
	PresencePenalty  float64 `json:"presence_penalty" jsonschema:"title=Presence Penalty,description=The presence penalty of the generation,required"`
}

// Provider defines the interface for AI providers
type Provider interface {
	// Generate returns a channel of events (tokens, tool calls, errors)
	Generate(ctx context.Context, params GenerationParams, messages []Message) <-chan Event

	// GenerateSync generates a complete response synchronously
	GenerateSync(ctx context.Context, params GenerationParams, messages []Message) (string, error)

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
