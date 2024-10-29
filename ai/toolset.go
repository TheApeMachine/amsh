package ai

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

// Toolset manages a collection of tools available to an agent
type Toolset struct {
	tools map[string]Tool
}

// NewToolset creates a new toolset and loads tools from configuration
func NewToolset() *Toolset {
	return &Toolset{
		tools: map[string]Tool{
			"memory":      tools.NewMemory(),
			"project":     tools.NewProject(),
			"helpdesk":    tools.NewHelpdesk(),
			"slack":       tools.NewSlack(),
			"browser":     tools.NewBrowser(),
			"environment": tools.NewEnvironment(),
		},
	}
}

func (toolset *Toolset) Use(name string, arguments map[string]any) string {
	if tool, ok := toolset.tools[name]; ok {
		return tool.Use(arguments)
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
