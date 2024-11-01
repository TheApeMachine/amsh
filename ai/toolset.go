package ai

import (
	"bytes"
	"context"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/tools"
)

// Toolset manages a collection of tools available to an agent
type Toolset struct {
	tools map[string]Tool
}

// NewToolset creates a new toolset and loads tools from configuration
func NewToolset(keys ...string) *Toolset {
	log.Info("Creating toolset", "keys", keys)

	var toolMap = map[string]Tool{
		"boards":      tools.NewBoards(),
		"browser":     tools.NewBrowser(),
		"environment": tools.NewEnvironment(),
		"helpdesk":    tools.NewHelpdesk(),
		"memory":      tools.NewMemory(),
		"neo4j":       tools.NewNeo4j(),
		"qdrant":      tools.NewQdrant("amsh", 1536),
		"slack":       tools.NewSlack(),
		"wiki":        tools.NewWiki(),
	}

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
	buf := new(bytes.Buffer)

	for _, tool := range toolset.tools {
		buf.WriteString(tool.GenerateSchema())
	}

	return buf.String()
}
