package tools

import (
	"context"
	"fmt"

	"github.com/theapemachine/amsh/ai/types"
)

// SetStateTool handles changing an agent's state
func SetStateTool(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get the state parameter
	stateStr, ok := args["state"].(string)
	if !ok {
		return "", fmt.Errorf("state parameter is required")
	}

	// Validate the state
	state := types.AgentState(stateStr)
	switch state {
	case types.StateIdle, types.StateThinking, types.StateWorking,
		types.StateWaiting, types.StateReviewing, types.StateDone:
		// Valid state
	default:
		return "", fmt.Errorf("invalid state: %s", stateStr)
	}

	// Get the agent from context
	agent, ok := ctx.Value("agent").(types.AgentContext)
	if !ok {
		return "", fmt.Errorf("agent not found in context")
	}

	// Set the new state
	agent.SetState(state)

	return fmt.Sprintf("State successfully updated to: %s", state), nil
}
