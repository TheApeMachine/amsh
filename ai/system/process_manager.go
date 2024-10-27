package system

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

/*
ProcessManager handles the orchestration of different processes across teams.
It maps process keys to their configurations and manages their execution.
*/
type ProcessManager struct {
	arch      *Architecture
	processes map[string]*Process
	mu        sync.RWMutex
}

// ProcessConfig holds the configuration for a specific process
type ProcessConfig struct {
	Team    string
	Process string
	Prompt  string
}

// Process represents a registered process with its configuration
type Process struct {
	Name        string
	Description string
	Teams       []string
	Handler     func(ctx context.Context, input interface{}) (interface{}, error)
}

// NewProcessManager creates a new process manager
func NewProcessManager(arch *Architecture) *ProcessManager {
	return &ProcessManager{
		arch:      arch,
		processes: make(map[string]*Process),
	}
}

/*
getProcessConfig retrieves the process configuration from the YAML config.
Returns team name, process prompt, and any error.
*/
func (pm *ProcessManager) getProcessConfig(processKey string) (*ProcessConfig, error) {
	// Map process keys to their team and config path
	processMap := map[string]ProcessConfig{
		"code_review": {
			Team:    "engineering",
			Process: "code_review",
		},
		"helpdesk_labelling": {
			Team:    "operations",
			Process: "helpdesk.labelling",
		},
		"discussion": {
			Team:    "management",
			Process: "discussion",
		},
		"backlog_refinement": {
			Team:    "management",
			Process: "backlog_refinement",
		},
		"sprint_planning": {
			Team:    "management",
			Process: "sprint_planning",
		},
		"retrospective": {
			Team:    "management",
			Process: "retrospective",
		},
	}

	config, exists := processMap[processKey]
	if !exists {
		return nil, fmt.Errorf("unknown process: %s", processKey)
	}

	// Get the process prompt from config
	prompt := viper.GetString(fmt.Sprintf("ai.setups.marvin.processes.%s", config.Process))
	if prompt == "" {
		return nil, fmt.Errorf("process prompt not found: %s", processKey)
	}

	config.Prompt = prompt
	return &config, nil
}

/*
HandleProcess is the unified entry point for handling any process.
It handles the routing to appropriate teams and agents based on the process key.
*/
func (pm *ProcessManager) HandleProcess(ctx context.Context, processKey string, input interface{}) (interface{}, error) {
	// Get process configuration
	config, err := pm.getProcessConfig(processKey)
	if err != nil {
		return nil, err
	}

	// Prepare assignment tool args
	args := map[string]interface{}{
		"team":    config.Team,
		"process": config.Process,
	}

	// Get the assignment tool
	assignmentTool := tools.AssignmentTool

	// Assign the process to the team
	_, err = assignmentTool(ctx, args)
	if err != nil {
		errnie.Error(err)
		return nil, err
	}

	// Get the team
	team := pm.arch.GetTeam(config.Team)
	if team == nil {
		return nil, fmt.Errorf("team not found: %s", config.Team)
	}

	// Get the teamlead
	teamlead := team.GetAgent("teamlead")
	if teamlead == nil {
		return nil, fmt.Errorf("teamlead not found for team: %s", config.Team)
	}

	// Create the process message with the configuration prompt
	processMsg := fmt.Sprintf("Process: %s\nPrompt: %s\nInput: %v",
		processKey,
		config.Prompt,
		input,
	)

	// Send to teamlead for processing
	teamlead.ReceiveMessage(processMsg)
	response, err := teamlead.ExecuteTask()
	if err != nil {
		return nil, err
	}

	return response, nil
}

// RegisterProcess registers a new process with the manager
func (pm *ProcessManager) RegisterProcess(name, description string, teams []string, handler func(ctx context.Context, input interface{}) (interface{}, error)) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if process already exists
	if _, exists := pm.processes[name]; exists {
		return fmt.Errorf("process %s already registered", name)
	}

	// Register the new process
	pm.processes[name] = &Process{
		Name:        name,
		Description: description,
		Teams:       teams,
		Handler:     handler,
	}

	return nil
}

// StartProcess executes a registered process with the given input
func (pm *ProcessManager) StartProcess(ctx context.Context, name string, input interface{}) (interface{}, error) {
	pm.mu.RLock()
	process, exists := pm.processes[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process %s not found", name)
	}

	// Execute the process handler
	return process.Handler(ctx, input)
}
