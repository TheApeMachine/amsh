package mastercomputer

import (
	"github.com/spf13/viper"
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
