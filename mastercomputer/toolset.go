package mastercomputer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/container"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/integration/boards"
	"github.com/theapemachine/amsh/integration/comms"
	"github.com/theapemachine/amsh/integration/git"
	"github.com/theapemachine/amsh/integration/trengo"
	"github.com/theapemachine/amsh/mastercomputer/memory"
)

// Toolset represents a set of tools available to a worker.
type Toolset struct {
	toolMap   map[string]openai.ChatCompletionToolParam
	toolDefs  map[string]*Tool // Added Tool Definitions map
	baseTools []string
	workloads map[string][]string
	events    *Events
}

// NewToolset returns a toolset based on the workload of the worker.
func NewToolset() *Toolset {
	errnie.Trace()
	// Initialize sideEffects within the NewToolset function to avoid initialization cycles
	sideEffects := map[string]func(parameters map[string]any) SideEffect{
		"set_state":   func(parameters map[string]any) SideEffect { return NewStateManager().SetState(parameters) },
		"memory":      func(parameters map[string]any) SideEffect { return memory.NewProxy(parameters) },
		"inspect":     func(parameters map[string]any) SideEffect { return NewInspector(parameters) },
		"assignment":  func(parameters map[string]any) SideEffect { return NewAssignment(parameters) },
		"worker":      func(parameters map[string]any) SideEffect { return NewWorker(parameters) },
		"work_items":  func(parameters map[string]any) SideEffect { return boards.NewProxy(parameters) },
		"slack":       func(parameters map[string]any) SideEffect { return comms.NewProxy(parameters) },
		"tweak":       func(parameters map[string]any) SideEffect { return NewTweaker(parameters) },
		"browser":     func(parameters map[string]any) SideEffect { return NewBrowser(parameters) },
		"helpdesk":    func(parameters map[string]any) SideEffect { return trengo.NewProxy(parameters) },
		"github":      func(parameters map[string]any) SideEffect { return git.NewProxy(parameters) },
		"environment": func(parameters map[string]any) SideEffect { return container.NewEnvironment(parameters) },
	}

	// Initialize toolMap and toolDefs separately to avoid duplicate field names
	toolMap := make(map[string]openai.ChatCompletionToolParam)
	toolDefs := make(map[string]*Tool)

	// Load configurations
	toolsets := viper.GetStringMapStringSlice("toolsets")
	toolsConfig := viper.GetStringMap("tools")

	// Initialize workloads map
	workloads := make(map[string][]string)

	// Store the toolset configurations
	for workload, tools := range toolsets {
		workloads[workload] = tools
	}

	// Create tools from configuration
	for toolName, toolInterface := range toolsConfig {
		if toolDef, ok := toolInterface.(map[string]interface{}); ok {
			description, descOk := toolDef["description"].(string)
			parameters, paramOk := toolDef["parameters"].(map[string]interface{})

			if descOk && paramOk {
				tool := NewTool(toolName, description, parameters, sideEffects[toolName])
				toolMap[toolName] = tool.Schema()
				toolDefs[toolName] = tool
			}
		}
	}

	// Define base tools that should always be available
	baseTools := []string{"memory", "inspect", "assignment"}

	return &Toolset{
		toolMap:   toolMap,
		toolDefs:  toolDefs,
		baseTools: baseTools,
		workloads: workloads,
		events:    NewEvents(),
	}
}

// WithBaseTools adds the common tools, such as memory management, and the ability to change state.
func (toolset *Toolset) WithBaseTools(in []openai.ChatCompletionToolParam) (out []openai.ChatCompletionToolParam) {
	errnie.Trace()

	for _, toolName := range toolset.baseTools {
		if tool, ok := toolset.toolMap[toolName]; ok {
			out = append(out, tool)
		}
	}

	return append(out, in...)
}

// Assign assigns a set of tools to a worker based on the workload.
func (toolset *Toolset) Assign(workload string) []openai.ChatCompletionToolParam {
	errnie.Trace()
	var tools []openai.ChatCompletionToolParam

	// First add base tools
	for _, toolName := range toolset.baseTools {
		if tool, ok := toolset.toolMap[toolName]; ok {
			tools = append(tools, tool)
		}
	}

	// Then add workload-specific tools
	if workloadTools, ok := toolset.workloads[workload]; ok {
		for _, toolName := range workloadTools {
			if tool, ok := toolset.toolMap[toolName]; ok {
				tools = append(tools, tool)
			}
		}
	}

	// Ensure we're not returning an empty array
	if len(tools) == 0 {
		// Add at least one default tool if no tools were assigned
		if tool, ok := toolset.toolMap["memory"]; ok {
			tools = append(tools, tool)
		}
	}

	return tools
}

// Use utilizes a tool based on the tool call passed in.
func (toolset *Toolset) Use(
	toolCall openai.ChatCompletionMessageToolCall, workerID string,
) (out openai.ChatCompletionToolMessageParam) {
	errnie.Trace()

	args := map[string]any{}
	var err error

	if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return openai.ToolMessage(toolCall.ID, err.Error())
	}

	msg := []string{
		"WORKER: " + workerID,
		"TOOL  : " + toolCall.Function.Name,
	}

	for k, v := range args {
		msg = append(msg, fmt.Sprintf("%s: %v", k, v))
	}

	// Emit an event for tool call
	toolset.events.channel <- Event{
		Timestamp: time.Now(),
		Type:      "ToolCall",
		Message:   strings.Join(msg, "\n"),
		WorkerID:  workerID,
	}

	// Check if the tool exists
	if tool, exists := toolset.toolDefs[toolCall.Function.Name]; exists {
		// Use the tool's Use method
		response := tool.Use(toolCall, workerID)

		// Emit an event for tool response
		toolset.events.channel <- Event{
			Timestamp: time.Now(),
			Type:      "ToolResponse",
			Message:   response,
			WorkerID:  workerID,
		}

		return openai.ToolMessage(toolCall.ID, response)
	}

	return openai.ToolMessage(toolCall.ID, "Tool not implemented")
}
