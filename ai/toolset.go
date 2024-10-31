package ai

import (
	"context"
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

var toolMap = map[string]Tool{
	"boards":      tools.NewBoards(),
	"browser":     tools.NewBrowser(),
	"environment": tools.NewEnvironment(),
	"helpdesk":    tools.NewHelpdesk(),
	"memory":      tools.NewMemory(),
	"slack":       tools.NewSlack(),
	"wiki":        tools.NewWiki(),
}

// Toolset manages a collection of tools available to an agent
type Toolset struct {
	tools map[string]Tool
}

// NewToolset creates a new toolset and loads tools from configuration
func NewToolset(keys ...string) *Toolset {
	tools := make(map[string]Tool)

	for _, key := range keys {
		tools[key] = toolMap[key]
	}

	return &Toolset{
		tools: tools,
	}
}

func (toolset *Toolset) Use(ctx context.Context, name string, arguments map[string]any) string {
	if tool, ok := toolset.tools[name]; ok {
		return tool.Use(ctx, arguments)
	}

	return ""
}

func (toolset *Toolset) GetTool(name string) (Tool, bool) {
	if tool, ok := toolset.tools[name]; ok {
		return tool, true
	}

	return nil, false
}

func (toolset *Toolset) Schemas() string {
	schema := jsonschema.Reflect(toolset.tools)
	buf, err := json.MarshalIndent(schema, "", "  ")

	if err != nil {
		errnie.Error(err)
	}

	return string(buf)
}
