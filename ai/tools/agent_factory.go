package tools

import (
	"context"
	"fmt"

	"github.com/theapemachine/amsh/ai/types"
)

/*
AgentFactoryTool allows dynamic creation of specialized agents.
*/
func AgentFactoryTool(ctx context.Context, args map[string]interface{}) (string, error) {
	// Extract required parameters
	role, ok := args["role"].(string)
	if !ok {
		return "", fmt.Errorf("role parameter is required")
	}

	specialization, ok := args["specialization"].(string)
	if !ok {
		return "", fmt.Errorf("specialization parameter is required")
	}

	// Get the agent manager from context
	agentManager, ok := ctx.Value("agents").(types.AgentManager)
	if !ok {
		return "", fmt.Errorf("agent manager not found in context")
	}

	// Create agent configuration
	config := types.AgentConfig{
		Role:           role,
		Specialization: specialization,
		SystemPrompt:   fmt.Sprintf("You are an AI agent specialized in %s. %s", role, specialization),
		UserPrompt:     fmt.Sprintf("Use your %s expertise to help process user requests", role),
	}

	// Create the agent
	agent, err := agentManager.CreateAgent(config)
	if err != nil {
		return "", fmt.Errorf("failed to create agent: %w", err)
	}

	return fmt.Sprintf("Successfully created agent with ID %s", agent.GetID()), nil
}
