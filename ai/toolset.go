package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"

	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

var Delivery = map[string]*Agent{}

// Toolset manages a collection of tools available to an agent
type Toolset struct {
	tools map[string]Tool
}

// NewToolset creates a new toolset and loads tools from configuration
func NewToolset(keys ...string) *Toolset {
	errnie.Info("Creating toolset %v", keys)

	var toolMap = map[string]Tool{
		"boards":      tools.NewBoards(),
		"browser":     tools.NewBrowser(),
		"environment": tools.NewEnvironment(),
		"recruit":     tools.NewRecruit(),
		"github":      tools.NewGithub(),
		"helpdesk":    tools.NewHelpdesk(),
		"neo4j":       tools.NewNeo4j(),
		"qdrant":      tools.NewQdrant("amsh", 1536),
		"slack":       tools.NewSlack(),
		"wiki":        tools.NewWiki(),
		"inspect":     tools.NewInspect(),
		"workload":    tools.NewWorkload(),
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
	errnie.Info("using tool %s", name)

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

	for name, tool := range toolset.tools {
		if tool == nil {
			errnie.Warn("tool %s is nil", name)
			continue
		}

		buf.WriteString(tool.GenerateSchema())
	}

	return buf.String()
}

func ExecuteToolCalls(agent *Agent, accumulator string) string {
	// Extract all Markdown JSON blocks.
	pattern := regexp.MustCompile("(?s)```json\\s*([\\s\\S]*?)```")
	matches := pattern.FindAllStringSubmatch(accumulator, -1)

	// To get the tool that was used, we need to unmarshal the JSON string.
	for _, match := range matches {
		var data map[string]any
		errnie.MustVoid(json.Unmarshal([]byte(match[1]), &data))

		if toolValue, ok := data["tool_name"].(string); ok {
			return agent.Toolset.Use(agent.ctx, toolValue, data)
		}
	}

	return "all tool calls executed"
}
