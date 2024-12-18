package ai

import (
	"context"
	"io"
)

// Tool represents a capability that can be used by an agent
type Tool interface {
	GenerateSchema() string
	Use(ctx context.Context, args map[string]any) string
}

// InteractiveTool represents a tool that requires ongoing IO communication
type InteractiveTool interface {
	Tool
	GetIO() io.ReadWriteCloser
	IsInteractive() bool
}
