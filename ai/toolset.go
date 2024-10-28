package ai

import (
	"context"
	"fmt"
	"sync"

	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/ai/types"
)

// Toolset manages a collection of tools available to an agent
type Toolset struct {
	tools    map[string]types.Tool
	mu       sync.RWMutex
	defaults []string
	roles    map[string][]string
}

// NewToolset creates a new toolset and loads tools from configuration
func NewToolset() *Toolset {
	ts := &Toolset{
		tools:    make(map[string]types.Tool),
		defaults: []string{},
		roles:    make(map[string][]string),
	}

	return ts
}

func (toolset *Toolset) GetTool(name string) (types.Tool, error) {
	toolset.mu.RLock()
	defer toolset.mu.RUnlock()

	switch name {
	case "browser":
		return tools.NewBrowser(), nil
	case "environment":
		return tools.NewEnvironmentTool()
	case "work_items":
		return tools.NewWorkItemsTool(context.Background())
	case "slack":
		return tools.NewSlackTool()
	default:
		return nil, fmt.Errorf("tool %s not found", name)
	}
}
