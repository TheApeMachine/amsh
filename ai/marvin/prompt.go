package marvin

import (
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
)

type Prompt struct {
	systemPrompt string
	rolePrompt   string
	userPrompt   string
	processes    []Process
}

func NewPrompt(role string) *Prompt {
	return &Prompt{
		systemPrompt: viper.GetViper().GetString("ai.setups.marvin.templates.system"),
		rolePrompt:   viper.GetViper().GetString("ai.setups.marvin.templates." + role),
	}
}

func (prompt *Prompt) SetUserPrompt(userPrompt string) {
	prompt.userPrompt = userPrompt
}

func (prompt *Prompt) System() provider.Message {
	processes := []string{}

	for _, process := range prompt.processes {
		processes = append(processes, process.GenerateSchema())
	}

	return provider.Message{
		Role: "system",
		Content: utils.JoinWith(
			"\n\n",
			prompt.systemPrompt,
			prompt.rolePrompt,
			"The following schema determines the manner in which you should respond to the incoming context, as well as how to structure your JSON response.",
			utils.JoinWith("\n",
				"<schema>",
				utils.JoinWith("\n\n", processes...),
				"</schema>",
			),
			"You should format your response as a valid JSON object, using the schema as a guide.",
			"You should go into as much detail as possible, and include all relevant information.",
			"You should not focus on a final answer, but rather focus on the most detailed execution of the schema, using the context as a source of information.",
			"You should never make any claims about the context, and never provide baseless conclusions, always focus only on the facts, no matter how confident you are.",
			"You should always fullfil the schema, in the most detailed way possible, no matter what.",
		),
	}
}

func (prompt *Prompt) User() provider.Message {
	return provider.Message{
		Role:    "user",
		Content: prompt.userPrompt,
	}
}

func (prompt *Prompt) Context() provider.Message {
	return provider.Message{
		Role: "assistant",
		Content: utils.JoinWith("\n\n",
			utils.JoinWith("\n",
				"<context>",
				prompt.userPrompt,
				"</context>",
			),
			"Please respond according to the schema provided, and do not deviate from it.",
			"Your response should be a valid JSON object, encapsulated by the <response> and </response> tags.",
		),
	}
}

func (prompt *Prompt) AddProcess(process Process) {
	prompt.processes = append(prompt.processes, process)
}
