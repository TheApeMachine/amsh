package ai

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
	"github.com/theapemachine/amsh/utils"
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

/*
initializeAgents initializes the agents for the team from the configuration file.
*/
func (team *Team) initializeAgents() error {
	agentRoles := viper.GetStringSlice("ai.setups.marvin.teams." + team.Name)

	for _, role := range agentRoles {
		team.mu.Lock()
		team.Agents[role] = NewAgent(
			utils.NewName(),
			types.Role(role),
			viper.GetString(fmt.Sprintf("ai.setups.marvin.system")),
			viper.GetString(fmt.Sprintf("ai.setups.marvin.teams.%s.%s", team.Name, role)),
			team.Toolset.GetToolsForRole(string(types.Role(role))),
			provider.NewRandomProvider(map[string]string{
				"openai":    os.Getenv("OPENAI_API_KEY"),
				"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
				"gemini":    os.Getenv("GOOGLE_API_KEY"),
				"cohere":    os.Getenv("COHERE_API_KEY"),
			}),
		)
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
