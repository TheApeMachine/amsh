package ai

import (
	"fmt"
	"sync"

	"github.com/theapemachine/amsh/ai/types"
)

// Team represents a group of agents working together
type Team struct {
	agents  map[string][]*Agent // Changed from Role to string
	toolset *Toolset
	mu      sync.RWMutex
}

// NewTeam creates a new team instance
func NewTeam(toolset *Toolset) *Team {
	return &Team{
		agents:  make(map[string][]*Agent),
		toolset: toolset,
	}
}

// GetNextAgentID generates a unique ID for a new agent
func (t *Team) GetNextAgentID() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Count total agents across all roles
	total := 0
	for _, agents := range t.agents {
		total += len(agents)
	}

	return total + 1
}

// AddAgent adds a new agent to the team
func (t *Team) AddAgent(agent *Agent) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if agent == nil {
		return fmt.Errorf("cannot add nil agent")
	}

	role := string(agent.GetRole()) // Convert types.Role to string for map key

	// Initialize the role slice if it doesn't exist
	if _, exists := t.agents[role]; !exists {
		t.agents[role] = make([]*Agent, 0)
	}

	// Add the agent to the appropriate role slice
	t.agents[role] = append(t.agents[role], agent)
	return nil
}

// GetAgentsByRole returns all agents with the specified role
func (t *Team) GetAgentsByRole(role types.Role) []*Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.agents[string(role)]
}

// GetAllAgents returns all agents in the team
func (t *Team) GetAllAgents() []*Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var allAgents []*Agent
	for _, agents := range t.agents {
		allAgents = append(allAgents, agents...)
	}
	return allAgents
}

// GetToolset returns the team's toolset
func (t *Team) GetToolset() *Toolset {
	return t.toolset
}

// Shutdown gracefully shuts down all agents in the team
func (t *Team) Shutdown() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, agents := range t.agents {
		for _, agent := range agents {
			agent.Shutdown()
		}
	}

	// Clear the agents map
	t.agents = make(map[string][]*Agent)
}

// GetMessageCount returns the total number of messages processed by the team
func (t *Team) GetMessageCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var count int
	for _, agents := range t.agents {
		for _, agent := range agents {
			count += agent.GetMessageCount()
		}
	}
	return count
}
