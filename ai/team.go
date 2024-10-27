package ai

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
)

// Team represents a group of AI agents working together
type Team struct {
	toolset  *Toolset
	agents   map[string]*Agent
	provider provider.Provider
	mu       sync.RWMutex
}

// NewTeam creates a new team with the given toolset
func NewTeam(toolset *Toolset) *Team {
	team := &Team{
		toolset: toolset,
		agents:  make(map[string]*Agent),
	}

	// Initialize agents from config
	if err := team.initializeAgents(); err != nil {
		// Log error but continue with empty team
		fmt.Printf("warning: failed to initialize agents: %v\n", err)
	}

	return team
}

func (t *Team) initializeAgents() error {
	// Get agent configurations from viper
	agentRoles := viper.GetStringMap("ai.prompt")

	for role := range agentRoles {
		if role == "system" || role == "processes" {
			continue // Skip system prompt and processes
		}

		// Get role-specific configuration
		roleConfig := viper.GetString(fmt.Sprintf("ai.prompt.%s.role", role))
		if roleConfig == "" {
			continue
		}

		// Get system prompt and replace placeholders
		systemPrompt := viper.GetString("ai.prompt.system")
		systemPrompt = fmt.Sprintf(systemPrompt,
			map[string]interface{}{
				"name":            role,
				"role":            role,
				"job_description": roleConfig,
			},
		)

		// Map config roles to types.Role
		var agentRole types.Role
		switch role {
		case "researcher":
			agentRole = types.RoleResearcher
		case "analyst":
			agentRole = types.RoleAnalyst
		default:
			// Skip roles that aren't defined in types.Role
			continue
		}

		// Create agent with configuration from YAML
		agent := NewAgent(
			role,         // id
			agentRole,    // role (using predefined type)
			systemPrompt, // system prompt
			roleConfig,   // user prompt/job description
			t.toolset,    // toolset
			nil,          // provider will be set later
		)

		t.mu.Lock()
		t.agents[role] = agent
		t.mu.Unlock()
	}

	return nil
}

// GetAgent returns an agent by role
func (t *Team) GetAgent(role string) *Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.agents[role]
}

// GetResearcher returns the research agent
func (t *Team) GetResearcher() *Agent {
	return t.GetAgent("researcher")
}

// GetAnalyst returns the analyst agent
func (t *Team) GetAnalyst() *Agent {
	return t.GetAgent("analyst")
}

// SetProvider sets the LLM provider for the team
func (t *Team) SetProvider(p provider.Provider) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.provider = p
	for _, agent := range t.agents {
		if agent != nil {
			agent.provider = p
		}
	}
}

// Shutdown performs cleanup for all team members
func (t *Team) Shutdown() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, agent := range t.agents {
		if agent != nil {
			agent.Shutdown()
		}
	}
}
