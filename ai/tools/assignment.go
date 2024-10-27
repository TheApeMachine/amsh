package tools

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/types"
)

// IsValidRole checks if a role is valid according to the configuration
func IsValidRole(role string) bool {
	validRoles := []string{
		"sequencer",
		"recruiter",
		"planner",
		"prompt_engineer",
		"reasoner",
		"researcher",
		"teamlead",
		"actor",
	}

	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

// IsValidProcess checks if a process exists in the configuration
func IsValidProcess(process string) bool {
	processes := viper.GetStringMap("ai.setups.marvin.processes")
	_, exists := processes[process]
	return exists
}

// GetProcessPrompt retrieves the process prompt from configuration
func GetProcessPrompt(process string) string {
	return viper.GetString(fmt.Sprintf("ai.setups.marvin.processes.%s", process))
}

// AssignmentTool handles assigning workloads to teams and processes
func AssignmentTool(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get the required parameters
	teamName, ok := args["team"].(string)
	if !ok {
		return "", fmt.Errorf("team parameter is required")
	}

	process, ok := args["process"].(string)
	if !ok {
		return "", fmt.Errorf("process parameter is required")
	}

	// Validate the process exists
	if !IsValidProcess(process) {
		return "", fmt.Errorf("invalid process: %s", process)
	}

	// Get the team manager from context
	teamManager, ok := ctx.Value("team").(types.TeamManager)
	if !ok {
		return "", fmt.Errorf("team manager not found in context")
	}

	// Get the team
	team, err := teamManager.GetTeam(teamName)
	if err != nil {
		return "", fmt.Errorf("failed to find team %s: %w", teamName, err)
	}
	if team == nil {
		return "", fmt.Errorf("team not found: %s", teamName)
	}

	return fmt.Sprintf("Successfully assigned process %s to team %s", process, teamName), nil
}
