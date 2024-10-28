package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
	"github.com/theapemachine/amsh/utils"
)

// Team represents a group of AI agents working together
type Team struct {
	Name       string            `json:"name"`
	Agents     map[string]*Agent `json:"agents"`
	TeamLead   *Agent            `json:"teamlead"` // Added explicit teamlead field
	Toolset    *Toolset          `json:"toolset"`
	Process    *Process          `json:"process"`
	mu         sync.RWMutex      `json:"-"`
	completion chan bool         `json:"-"`
}

// NewTeam creates a new team with the given toolset
func NewTeam(toolset *Toolset) *Team {
	// Create the teamlead first
	teamlead := NewAgent(
		utils.NewName(),
		"teamlead",
		viper.GetString("ai.setups.marvin.agents.teamlead.role"),
		viper.GetString("ai.setups.marvin.agents.teamlead.specialization"),
		toolset.GetToolsForRole("teamlead"),
		provider.NewRandomProvider(map[string]string{
			"openai":    os.Getenv("OPENAI_API_KEY"),
			"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
			"google":    os.Getenv("GOOGLE_API_KEY"),
			"cohere":    os.Getenv("COHERE_API_KEY"),
		}),
	)

	return &Team{
		Toolset:    toolset,
		Agents:     make(map[string]*Agent),
		TeamLead:   teamlead,
		completion: make(chan bool),
	}
}

// ExecuteProcess now uses the teamlead to orchestrate
func (team *Team) ExecuteProcess(ctx context.Context, process *Process) <-chan []byte {
	log.Info("Executing process", "process", process.Name)
	outputChan := make(chan []byte)

	go func() {
		defer close(outputChan)

		// Initialize Process if nil
		if process == nil {
			outputChan <- []byte("Error: nil process provided")
			return
		}

		team.Process = process
		team.Process.Status = ProcessStatusRunning

		// Initialize Steps if nil
		if team.Process.Steps == nil {
			outputChan <- []byte("Error: no steps defined in process")
			return
		}

		for i := range team.Process.Steps {
			select {
			case <-ctx.Done():
				return
			default:
				step := &team.Process.Steps[i]
				team.Process.CurrentStep = i

				// Ensure TeamLead exists
				if team.TeamLead == nil {
					outputChan <- []byte("Error: no team lead assigned")
					return
				}

				// Ensure assignment tool exists
				assignmentTool := team.TeamLead.Tools["assignment"]
				if assignmentTool == nil {
					outputChan <- []byte("Error: team lead missing assignment tool")
					return
				}

				assignmentResult, err := assignmentTool.Execute(ctx, map[string]interface{}{
					"step": step.Name,
					"team": team,
				})
				if err != nil {
					outputChan <- []byte(fmt.Sprintf("Error in team assignment: %v", err))
					continue
				}

				assignedAgent := team.parseAssignment(assignmentResult)
				if assignedAgent == nil {
					outputChan <- []byte(fmt.Sprintf("No agent assigned for step: %s", step.Name))
					continue
				}

				result, err := assignedAgent.ExecuteTask()
				if err != nil {
					outputChan <- []byte(fmt.Sprintf("Error in step %s: %v", step.Name, err))
					continue
				}

				step.Output = result
				step.Status = StepStatusCompleted
				step.EndTime = time.Now()

				outputChan <- []byte(result)
			}
		}

		team.Process.Status = ProcessStatusComplete
		// Only send completion signal if channel exists
		if team.completion != nil {
			team.completion <- true
		}
	}()

	return outputChan
}

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
	log.Info("Adding member to team", "agent", agent.GetID())
	team.mu.Lock()
	defer team.mu.Unlock()

	if team.Agents == nil {
		team.Agents = make(map[string]*Agent)
	}
	team.Agents[agent.GetID()] = agent
}

// GetAvailableAgentRoles returns a list of all agent roles currently in the team
func (team *Team) GetAvailableAgentRoles() []string {
	log.Info("Getting available agent roles")
	team.mu.RLock()
	defer team.mu.RUnlock()

	roles := make([]string, 0, len(team.Agents))
	for _, agent := range team.Agents {
		roles = append(roles, string(agent.GetRole()))
	}
	return roles
}

// GetAgentWithCapability returns an agent that has the specified capability
func (team *Team) GetAgentWithCapability(capability string) *Agent {
	log.Info("Getting agent with capability", "capability", capability)
	team.mu.RLock()
	defer team.mu.RUnlock()

	for _, agent := range team.Agents {
		// Check if agent has this capability
		if agent.HasCapability(capability) {
			return agent
		}
	}
	return nil
}

// Members returns a list of all agent IDs in the team
func (team *Team) Members() []string {
	team.mu.RLock()
	defer team.mu.RUnlock()

	members := make([]string, 0, len(team.Agents))
	for id := range team.Agents {
		members = append(members, id)
	}
	return members
}

// CreateTeamForTask dynamically assembles a team based on task requirements
func (team *Team) CreateTeamForTask(ctx context.Context, task string) error {
	log.Info("Creating team for task")

	// First, analyze the task to determine required capabilities
	requiredCapabilities := team.analyzeTaskRequirements(task)
	log.Info("Identified required capabilities", "capabilities", requiredCapabilities)

	// Create specialized agents based on identified needs
	for capability := range requiredCapabilities {
		agentID := fmt.Sprintf("%s_specialist", capability)

		// Create agent with specific specialization
		agent := NewAgent(
			agentID,
			types.Role(capability),
			fmt.Sprintf("You are an AI agent specialized in %s. Your task is to help answer: %s", capability, task),
			fmt.Sprintf("Use your expertise in %s to contribute to the team's understanding", capability),
			team.Toolset.GetToolsForCapability(capability),
			nil, // Let the agent choose its provider based on capability
		)

		// Add agent to team
		team.mu.Lock()
		team.Agents[agentID] = agent
		team.mu.Unlock()

		log.Info("Added agent to team", "role", capability, "id", agentID)
	}

	return nil
}

// analyzeTaskRequirements determines what capabilities are needed for a task
func (team *Team) analyzeTaskRequirements(task string) map[string]float64 {
	log.Info("Analyzing task requirements")
	capabilities := make(map[string]float64)

	// Example analysis logic (this should be more sophisticated in practice)
	task = strings.ToLower(task)

	// Research capability for fact-checking
	if strings.Contains(task, "does") || strings.Contains(task, "own") {
		capabilities["researcher"] = 0.9
	}

	// Legal expertise for ownership questions
	if strings.Contains(task, "own") {
		capabilities["legal_expert"] = 0.8
	}

	// Scientific knowledge for space-related questions
	if strings.Contains(task, "moon") {
		capabilities["scientist"] = 0.7
	}

	// Historical knowledge for questions about historical figures
	if strings.Contains(task, "elvis") {
		capabilities["historian"] = 0.8
	}

	// Always include an analyst to synthesize information
	capabilities["analyst"] = 1.0

	return capabilities
}

// parseAssignment interprets the teamlead's assignment response and returns the appropriate agent
func (team *Team) parseAssignment(assignmentResult string) *Agent {
	log.Info("Parsing assignment", "assignment", assignmentResult)
	team.mu.RLock()
	defer team.mu.RUnlock()

	// Parse the assignment result to extract agent ID/role
	// This is a simple implementation - enhance based on your actual response format
	for _, agent := range team.Agents {
		if strings.Contains(strings.ToLower(assignmentResult), strings.ToLower(string(agent.GetRole()))) {
			return agent
		}
	}

	// If no specific agent found, return the teamlead as fallback
	return team.TeamLead
}
