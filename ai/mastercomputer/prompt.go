package mastercomputer

import (
	"strings"

	"github.com/spf13/viper"
)

type Prompt struct {
	key          string
	systemPrompt string
	rolePrompt   string
	buffer       strings.Builder
}

func NewPrompt(key, role string) *Prompt {
	return &Prompt{
		key:          key,
		systemPrompt: viper.GetViper().GetString("ai.setups." + key + ".templates.system"),
		rolePrompt:   viper.GetViper().GetString("ai.setups." + key + ".templates." + role),
	}
}
