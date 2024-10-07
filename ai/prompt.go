package ai

import (
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type Prompt struct {
	SessionID string   `json:"session_id"`
	System    []string `json:"system"`
	Assistant []string `json:"assistant"`
	Tool      []string `json:"tool"`
	Function  []string `json:"function"`
	User      []string `json:"user"`
}

func NewPrompt(agentType string) *Prompt {
	return &Prompt{
		SessionID: uuid.New().String(),
		System: []string{
			viper.GetViper().GetString(`ai.${agentType}.system`),
		},
		User: []string{
			viper.GetViper().GetString(`ai.${agentType}.user`),
		},
	}
}
