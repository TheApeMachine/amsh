package mastercomputer

import (
	"strings"

	"github.com/openai/openai-go"
	"github.com/spf13/viper"
)

type Task struct {
	role      string
	sysStr    string
	usrStr    string
	system    openai.ChatCompletionMessageParamUnion
	user      openai.ChatCompletionMessageParamUnion
	responses []map[string]openai.ChatCompletionMessageParamUnion
}

func NewTask(name, role, system, user string) *Task {
	system = strings.ReplaceAll(system, "{name}", name)
	system = strings.ReplaceAll(system, "{role}", role)
	system = strings.ReplaceAll(system, "{job_description}", viper.GetViper().GetString("ai.prompt."+role))

	user = "[USER PROMPT]\n" + user

	return &Task{
		role:      role,
		sysStr:    system + "\n\n",
		usrStr:    user,
		system:    openai.SystemMessage(system),
		user:      openai.UserMessage(user),
		responses: make([]map[string]openai.ChatCompletionMessageParamUnion, 0),
	}
}
