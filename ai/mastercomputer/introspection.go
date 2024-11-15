// introspection.go
package mastercomputer

import (
	"reflect"
	"sync"

	"github.com/theapemachine/amsh/ai"
)

// Capability represents a process or tool capability
type Capability struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "process" or "tool"
	Description string                 `json:"description"`
	Behaviors   []string               `json:"behaviors"`
	Inputs      map[string]string      `json:"inputs"`
	Outputs     map[string]string      `json:"outputs"`
	Examples    []string               `json:"examples"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// IntrospectionSystem manages capability discovery and validation
type IntrospectionSystem struct {
	mu           sync.RWMutex
	capabilities map[string]Capability
	toolset      *ai.Toolset
}

func NewIntrospectionSystem(toolset *ai.Toolset) *IntrospectionSystem {
	is := &IntrospectionSystem{
		capabilities: make(map[string]Capability),
		toolset:      toolset,
	}

	// Register built-in processes
	is.registerCoreProcesses()

	// Register available tools
	is.registerTools()

	return is
}

// RegisterCapability adds a new capability to the system
func (is *IntrospectionSystem) RegisterCapability(cap Capability) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.capabilities[cap.Name] = cap
}

// GetCapability retrieves information about a specific capability
func (is *IntrospectionSystem) GetCapability(name string) (Capability, bool) {
	is.mu.RLock()
	defer is.mu.RUnlock()
	cap, exists := is.capabilities[name]
	return cap, exists
}

func (is *IntrospectionSystem) MockCapability(cap Capability) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.capabilities[cap.Name] = cap
}

// ListProcesses returns all registered process capabilities
func (is *IntrospectionSystem) ListProcesses() []Capability {
	is.mu.RLock()
	defer is.mu.RUnlock()

	processes := make([]Capability, 0)
	for _, cap := range is.capabilities {
		if cap.Type == "process" {
			processes = append(processes, cap)
		}
	}
	return processes
}

// ListTools returns all registered tool capabilities
func (is *IntrospectionSystem) ListTools() []Capability {
	is.mu.RLock()
	defer is.mu.RUnlock()

	tools := make([]Capability, 0)
	for _, cap := range is.capabilities {
		if cap.Type == "tool" {
			tools = append(tools, cap)
		}
	}
	return tools
}

// ValidateBehavior checks if a behavior is valid for a given capability
func (is *IntrospectionSystem) ValidateBehavior(capName, behavior string) bool {
	cap, exists := is.GetCapability(capName)
	if !exists {
		return false
	}

	for _, b := range cap.Behaviors {
		if b == behavior {
			return true
		}
	}
	return false
}

// registerCoreProcesses sets up the built-in process types
func (is *IntrospectionSystem) registerCoreProcesses() {
	processes := []Capability{
		{
			Name:        "surface",
			Type:        "process",
			Description: "Immediate pattern analysis and structural relationships",
			Behaviors:   []string{"temporal", "spatial", "relational"},
			Inputs: map[string]string{
				"data":    "any",
				"context": "map[string]interface{}",
			},
			Outputs: map[string]string{
				"patterns":      "[]Pattern",
				"relationships": "[]Relationship",
			},
			Examples: []string{
				"surface<temporal> => next | cancel",
				"surface<spatial> => send | back",
			},
		},
		{
			Name:        "quantum",
			Type:        "process",
			Description: "Probabilistic state handling and superpositions",
			Behaviors:   []string{"superposition", "entanglement", "probability"},
			Inputs: map[string]string{
				"states":  "[]State",
				"context": "map[string]interface{}",
			},
			Outputs: map[string]string{
				"possibilities": "[]Possibility",
				"correlations":  "[]Correlation",
			},
			Examples: []string{
				"quantum<superposition> => next | cancel",
				"quantum<probability> => send | back",
			},
		},
		// Add other core processes...
	}

	for _, proc := range processes {
		is.RegisterCapability(proc)
	}
}

// registerTools analyzes and registers available tools
func (is *IntrospectionSystem) registerTools() {
	// Use reflection to analyze toolset
	toolsetType := reflect.TypeOf(is.toolset).Elem()

	for i := 0; i < toolsetType.NumField(); i++ {
		field := toolsetType.Field(i)

		// Create capability from tool
		cap := Capability{
			Name:        field.Name,
			Type:        "tool",
			Description: field.Tag.Get("description"),
			Behaviors:   []string{}, // Tools don't have behaviors
			Inputs:      make(map[string]string),
			Outputs:     make(map[string]string),
			Examples:    []string{},
		}

		// Analyze tool methods for inputs/outputs
		toolType := field.Type
		for j := 0; j < toolType.NumMethod(); j++ {
			method := toolType.Method(j)
			cap.Inputs[method.Name] = method.Type.String()
		}

		is.RegisterCapability(cap)
	}
}

// ContextPossibilities returns valid next steps in current context
func (is *IntrospectionSystem) ContextPossibilities(ctx ProgramContext) map[string][]string {
	possibilities := make(map[string][]string)

	// Determine valid next steps based on context
	possibilities["next_steps"] = is.getValidNextSteps(ctx)
	possibilities["valid_behaviors"] = is.getValidBehaviors(ctx)
	possibilities["error_handlers"] = []string{"next", "back", "cancel"}

	return possibilities
}

type ProgramContext struct {
	CurrentProcess string
	PreviousSteps  []string
	State          map[string]interface{}
}

func (is *IntrospectionSystem) getValidNextSteps(ctx ProgramContext) []string {
	_ = ctx
	// Implementation would consider:
	// - Current process type
	// - Previous steps
	// - State requirements
	return []string{} // Placeholder
}

func (is *IntrospectionSystem) getValidBehaviors(ctx ProgramContext) []string {
	_ = ctx
	// Implementation would consider:
	// - Current process capabilities
	// - Context requirements
	return []string{} // Placeholder
}
