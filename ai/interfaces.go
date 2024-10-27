package ai

import (
	"context"
)

// Provider defines the interface for AI providers
type Provider interface {
	Generate(ctx context.Context, messages []Message) (string, error)
	GenerateStream(ctx context.Context, messages []Message) <-chan string
}
