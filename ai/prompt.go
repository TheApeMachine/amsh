package ai

import (
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
Prompt represents a prompt for an AI agent.
*/
type Prompt struct {
	SessionID string   `json:"session_id"`
	System    []string `json:"system"`
	Assistant []string `json:"assistant"`
	Tool      []string `json:"tool"`
	Function  []string `json:"function"`
	User      []string `json:"user"`
}

/*
NewPrompt creates a new Prompt for a given agent type.
*/
func NewPrompt(agentType string) *Prompt {
	errnie.Debug("Creating new prompt for agent type: %s", agentType)
	system := viper.GetViper().GetString(`ai.`+agentType+`.system`) + "\n\n"
	user := viper.GetViper().GetString(`ai.`+agentType+`.user`) + "\n\n"

	errnie.Debug("[SYSTEM PROMPT]\n%s\n\n", system)
	errnie.Debug("[USER PROMPT]\n%s\n\n", user)

	return &Prompt{
		SessionID: uuid.New().String(),
		System: []string{
			system,
		},
		User: []string{
			user,
		},
	}
}
