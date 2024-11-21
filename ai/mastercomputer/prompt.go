package mastercomputer

import (
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
)

type Prompt struct {
	systemPrompt string
	rolePrompt   string
}

func NewPrompt(role string) *Prompt {
	return &Prompt{
		systemPrompt: viper.GetViper().GetString("ai.setups.mastercomputer.templates.system"),
		rolePrompt:   viper.GetViper().GetString("ai.setups.mastercomputer.templates." + role),
	}
}

func (prompt *Prompt) System() provider.Message {
	return provider.Message{
		Role:    "system",
		Content: utils.JoinWith("\n\n", prompt.systemPrompt, prompt.rolePrompt),
	}
}
