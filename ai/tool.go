package ai

import (
	"context"
)

// Tool interface defines the contract for all tools
type Tool interface {
	Name() string
	Use(ctx context.Context, args map[string]any) string
	GenerateSchema() string
}
