package ai

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/types"
)

// Team represents a group of AI agents working together
type Team struct {
	Name    string            `json:"name"`
	Agents  map[string]*Agent `json:"agents"`
	Toolset *Toolset          `json:"toolset"`
	mu      sync.RWMutex      `json:"-"`
}

// NewTeam creates a new team with the given toolset
func NewTeam(toolset *Toolset) *Team {
	team := &Team{
		Toolset: toolset,
		Agents:  make(map[string]*Agent),
	}

	// Initialize agents from config
	if err := team.initializeAgents(); err != nil {
		// Log error but continue with empty team
		fmt.Printf("warning: failed to initialize agents: %v\n", err)
	}

	return team
}

func (team *Team) initializeAgents() error {
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
			team.Toolset, // toolset
			nil,          // provider will be set later
		)

		team.mu.Lock()
		team.Agents[role] = agent
		team.mu.Unlock()
	}

	return nil
}

// GetAgent returns an agent by role
func (team *Team) GetAgent(role string) *Agent {
	team.mu.RLock()
	defer team.mu.RUnlock()
	return team.Agents[role]
}

// GetResearcher returns the research agent
func (team *Team) GetResearcher() *Agent {
	return team.GetAgent("researcher")
}

// GetAnalyst returns the analyst agent
func (team *Team) GetAnalyst() *Agent {
	return team.GetAgent("analyst")
}

// Shutdown performs cleanup for all team members
func (team *Team) Shutdown() {
	team.mu.RLock()
	defer team.mu.RUnlock()

	for _, agent := range team.Agents {
		if agent != nil {
			agent.Shutdown()
		}
	}
}

// AddMember adds a new agent to the team
func (team *Team) AddMember(agent *Agent) {
	team.mu.Lock()
	defer team.mu.Unlock()

	if team.Agents == nil {
		team.Agents = make(map[string]*Agent)
	}
	team.Agents[agent.GetID()] = agent
}
