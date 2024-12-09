package marvin

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/utils"
)

// Prompt manages the system and user prompts for an agent
type Prompt struct {
	role         string
	systemPrompt string
	rolePrompt   string
	userPrompt   string
	processes    []Process
}

// NewPrompt creates a new prompt with the given role
func NewPrompt(role string) *Prompt {
	return &Prompt{
		role:         role,
		systemPrompt: viper.GetString("ai.setups.marvin.templates.system"),
		rolePrompt:   viper.GetString("ai.setups.marvin.templates." + role),
		processes:    make([]Process, 0),
	}
}

// System returns the system prompt with injected schema
func (p *Prompt) System() provider.Message {
	// Generate schema block
	var schemaBlock string
	if len(p.processes) > 0 {
		schemas := make([]string, 0, len(p.processes))
		for _, process := range p.processes {
			schemas = append(schemas, process.GenerateSchema())
		}
		schemaBlock = utils.JoinWith("\n",
			"<schema>",
			utils.JoinWith("\n\n", schemas...),
			"</schema>",
		)
	}

	// Replace schema marker in templates if present
	systemPrompt := strings.Replace(p.systemPrompt, "{{schema}}", schemaBlock, 1)
	rolePrompt := strings.Replace(p.rolePrompt, "{{schema}}", schemaBlock, 1)

	return provider.Message{
		Role: "system",
		Content: utils.JoinWith("\n\n",
			systemPrompt,
			rolePrompt,
		),
	}
}

// User returns the user prompt
func (p *Prompt) User() provider.Message {
	return provider.Message{
		Role:    "user",
		Content: p.userPrompt,
	}
}

// Context returns the context prompt
func (p *Prompt) Context() provider.Message {
	return provider.Message{
		Role:    "system",
		Content: "",
	}
}

// SetUserPrompt sets the user prompt content
func (p *Prompt) SetUserPrompt(content string) {
	p.userPrompt = content
}

// AddProcess adds a process to the prompt
func (p *Prompt) AddProcess(process Process) {
	p.processes = append(p.processes, process)
}
