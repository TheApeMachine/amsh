package ai

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
Prompt uses composable template fragments, optionally with dynamic variables to
craft a new instruction to send to a Large Language Model so it understands a
given task, and is provided with any relevant context. Prompt fragments are
defined in the config file embedded in the binary, which is written to the
home directory of the user. It can be interacted with using Viper.
*/
type Prompt struct {
	queue   strings.Builder
	system  string
	modules string
	role    string
	context string
}

/*
NewPrompt creates an empty prompt that can be used to dynamically build sophisticated
instructions that represent a request for an AI to produce a specific response.
*/
func NewPrompt(role string) *Prompt {
	errnie.Debug("Creating prompt for %s with modules: %s", role, viper.GetViper().GetString(fmt.Sprintf("ai.modules.%s", role)))
	return &Prompt{
		queue:   strings.Builder{},
		system:  viper.GetViper().GetString("ai.prompt.system"),
		modules: viper.GetViper().GetString(fmt.Sprintf("ai.modules.%s", role)),
		role:    viper.GetViper().GetString(fmt.Sprintf("ai.prompt.profiles.%s", role)),
		context: "",
	}
}
