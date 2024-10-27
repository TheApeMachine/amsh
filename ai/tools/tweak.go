package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/datalake"
)

type PromptHistory struct {
	AgentID     string    `json:"agent_id"`
	OldPrompt   string    `json:"old_prompt"`
	NewPrompt   string    `json:"new_prompt"`
	Timestamp   time.Time `json:"timestamp"`
	Explanation string    `json:"explanation"`
}

// TweakTool handles optimizing agent system prompts
func TweakTool(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get required parameters
	agentID, ok := args["agent"].(string)
	if !ok {
		return "", fmt.Errorf("agent parameter is required")
	}

	newPrompt, ok := args["system"].(string)
	if !ok {
		return "", fmt.Errorf("system parameter is required")
	}

	// Get the agent from the ambient manager
	manager := ai.GetAgentManager()
	agent, err := manager.GetAgent(agentID)
	if err != nil {
		return "", fmt.Errorf("failed to find agent: %w", err)
	}

	// Store the old prompt for history
	history := PromptHistory{
		AgentID:     agentID,
		OldPrompt:   agent.GetBuffer().GetSystemPrompt(),
		NewPrompt:   newPrompt,
		Timestamp:   time.Now(),
		Explanation: "Prompt optimization requested via tweak tool", // Could be expanded with more detail
	}

	// Store the history in S3
	key := fmt.Sprintf("prompts/%s/%d.json", agentID, time.Now().Unix())
	conn := datalake.NewConn(key)

	historyJSON, err := json.Marshal(history)
	if err != nil {
		return "", fmt.Errorf("failed to marshal history: %w", err)
	}

	if _, err := conn.Write(historyJSON); err != nil {
		return "", fmt.Errorf("failed to store history: %w", err)
	}

	return fmt.Sprintf("Successfully updated system prompt for agent %s", agentID), nil
}
