package tools

import (
	"context"
	"fmt"

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

// AssignmentTool handles assigning workloads to teams
func AssignmentTool(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get the required parameters
	role, ok := args["role"].(string)
	if !ok {
		return "", fmt.Errorf("role parameter is required")
	}

	// Validate the role exists
	if !IsValidRole(role) {
		return "", fmt.Errorf("invalid role: %s", role)
	}

	// Get the workload parameter but don't validate it yet
	_, ok = args["workload"].(string)
	if !ok {
		return "", fmt.Errorf("workload parameter is required")
	}

	// Get the team manager from context but don't use it yet
	if _, ok := ctx.Value("team").(types.TeamManager); !ok {
		return "", fmt.Errorf("team manager not found in context")
	}

	return fmt.Sprintf("Successfully validated role %s for assignment", role), nil
}
