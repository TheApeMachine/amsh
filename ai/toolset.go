package ai

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/types"
	"github.com/theapemachine/amsh/berrt"
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

	// Load tool configurations from viper
	berrt.Error("Toolset", ts.loadFromConfig())
	return ts
}

func (ts *Toolset) loadFromConfig() error {
	// Load toolsets configuration
	toolsets := viper.GetStringMapStringSlice("toolsets")
	for role, tools := range toolsets {
		ts.roles[role] = tools
		if role == "base" {
			ts.defaults = tools
		}
	}

	// Load individual tool configurations
	toolsCfg := viper.GetStringMap("tools")
	for name, config := range toolsCfg {
		toolConfig, ok := config.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid tool configuration for %s", name)
		}

		description, _ := toolConfig["description"].(string)
		parameters, _ := toolConfig["parameters"].(map[string]interface{})

		// Create a new tool with a default handler that returns "not implemented"
		tool := types.NewBaseTool(
			name,
			description,
			parameters,
			func(ctx context.Context, args map[string]interface{}) (string, error) {
				return fmt.Sprintf("Tool %s not implemented", name), nil
			},
		)

		ts.tools[name] = tool
	}

	// Register the memory tool handler
	if err := ts.RegisterToolHandler("memory", types.MemoryTool); err != nil {
		return fmt.Errorf("failed to register memory tool: %w", err)
	}

	return nil
}

// GetToolsForRole returns the tools assigned to a specific role
func (ts *Toolset) GetToolsForRole(role string) map[string]types.Tool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tools := make(map[string]types.Tool)

	// Add default tools first
	for _, name := range ts.defaults {
		if tool, exists := ts.tools[name]; exists {
			tools[name] = tool
		}
	}

	// Add role-specific tools
	if roleTools, exists := ts.roles[role]; exists {
		for _, name := range roleTools {
			if tool, exists := ts.tools[name]; exists {
				tools[name] = tool
			}
		}
	}

	return tools
}

// RegisterToolHandler registers a handler for an existing tool
func (ts *Toolset) RegisterToolHandler(name string, handler types.ToolHandler) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tool, exists := ts.tools[name]
	if !exists {
		return fmt.Errorf("tool %s not found", name)
	}

	if baseTool, ok := tool.(*types.BaseTool); ok {
		baseTool.Handler = handler
	} else {
		return fmt.Errorf("tool %s is not a BaseTool", name)
	}

	return nil
}

// GetTool returns a specific tool by name
func (ts *Toolset) GetTool(name string) (types.Tool, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tool, exists := ts.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

// ListTools returns a list of all available tool names
func (ts *Toolset) ListTools() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tools := make([]string, 0, len(ts.tools))
	for name := range ts.tools {
		tools = append(tools, name)
	}
	return tools
}

// ListRoles returns a list of all available roles
func (ts *Toolset) ListRoles() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	roles := make([]string, 0, len(ts.roles))
	for role := range ts.roles {
		roles = append(roles, role)
	}
	return roles
}

// ExecuteTool executes a specific tool by name with the given arguments
func (ts *Toolset) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	tool, err := ts.GetTool(name)
	if err != nil {
		return "", err
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		return "", fmt.Errorf("failed to execute tool %s: %w", name, err)
	}

	return result, nil
}

// HasTool checks if a tool exists in the toolset
func (ts *Toolset) HasTool(name string) bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	_, exists := ts.tools[name]
	return exists
}

// GetToolSchema returns the schema for a specific tool
func (ts *Toolset) GetToolSchema(name string) (types.ToolSchema, error) {
	tool, err := ts.GetTool(name)
	if err != nil {
		return types.ToolSchema{}, err
	}
	return tool.GetSchema(), nil
}

// GetDefaultTools returns the list of default tool names
func (ts *Toolset) GetDefaultTools() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	defaults := make([]string, len(ts.defaults))
	copy(defaults, ts.defaults)
	return defaults
}

// GetRoleTools returns the list of tool names for a specific role
func (ts *Toolset) GetRoleTools(role string) []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tools, exists := ts.roles[role]
	if !exists {
		return []string{}
	}

	roleTools := make([]string, len(tools))
	copy(roleTools, tools)
	return roleTools
}

// AddTool adds a new tool to the toolset
func (ts *Toolset) AddTool(name string, tool types.Tool) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if _, exists := ts.tools[name]; exists {
		return fmt.Errorf("tool %s already exists", name)
	}

	ts.tools[name] = tool
	return nil
}

// GetToolsForCapability returns tools that match a specific capability
func (ts *Toolset) GetToolsForCapability(capability string) map[string]types.Tool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// For now, treat capability same as role
	tools := make(map[string]types.Tool)
	for _, toolName := range ts.GetRoleTools(capability) {
		if tool, err := ts.GetTool(toolName); err == nil {
			tools[toolName] = tool
		}
	}
	return tools
}
