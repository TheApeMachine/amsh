package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/theapemachine/amsh/datalake"
)

type PromptHistory struct {
	AgentID   string    `json:"agent_id"`
	OldPrompt string    `json:"old_prompt"`
	NewPrompt string    `json:"new_prompt"`
	Timestamp time.Time `json:"timestamp"`
}

func TweakTool(ctx context.Context, args map[string]interface{}) (string, error) {
	agentID, ok := args["agent"].(string)
	if !ok {
		return "", fmt.Errorf("agent parameter must be a string")
	}

	newPrompt, ok := args["system"].(string)
	if !ok {
		return "", fmt.Errorf("system parameter must be a string")
	}

	// Get the agent from the ambient manager
	manager := GetAgentManager()
	agent, err := manager.GetAgent(agentID)
	if err != nil {
		return "", fmt.Errorf("failed to find agent: %w", err)
	}

	// Create prompt history record
	history := PromptHistory{
		AgentID:   agentID,
		OldPrompt: agent.context,
		NewPrompt: newPrompt,
		Timestamp: time.Now(),
	}

	// Store the history in S3
	conn := datalake.NewConn(&sync.WaitGroup{}, "amsh")
	conn.SetKey(fmt.Sprintf("prompts/%s/%d.json", agentID, time.Now().Unix()))

	historyJSON, err := json.Marshal(history)
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt history: %w", err)
	}

	if _, err := conn.Write(historyJSON); err != nil {
		return "", fmt.Errorf("failed to store prompt history: %w", err)
	}

	// Update the agent's prompt
	agent.context = newPrompt

	return fmt.Sprintf("Successfully updated prompt for agent %s", agentID), nil
}
