package ai

import (
	"fmt"
	"sync"

	"github.com/theapemachine/amsh/datalake"
)

// AgentManager handles agent lifecycle and state management
type AgentManager struct {
	agents  map[string]*Agent
	storage *datalake.Conn
	mu      sync.RWMutex
}

var (
	manager *AgentManager
	once    sync.Once
)

// GetAgentManager returns the singleton instance of AgentManager
func GetAgentManager() *AgentManager {
	once.Do(func() {
		manager = &AgentManager{
			agents:  make(map[string]*Agent),
			storage: datalake.NewConn(), // Or whatever storage implementation
		}
	})
	return manager
}

// RegisterAgent adds an agent to the manager
func (m *AgentManager) RegisterAgent(agent *Agent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agent.GetID()] = agent
}

// GetAgent retrieves an agent by ID
func (m *AgentManager) GetAgent(id string) (*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[id]

	if !exists {
		return nil, fmt.Errorf("agent %s not found", id)
	}

	return agent, nil
}
