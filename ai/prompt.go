package ai

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

/*
Prompt uses composable template fragments, optionally with dynamic variables to
craft a new instruction to send to a Large Language Model so it understands a
given task, and is provided with any relevant context. Prompt fragments are
defined in the config file embedded in the binary, which is written to the
home directory of the user. It can be interacted with using Viper.
*/
type Prompt struct {
	queue strings.Builder
}

/*
NewPrompt creates an empty prompt that can be used to dynamically build sophisticated
instructions that represent a request for an AI to produce a specific response.
*/
func NewPrompt() *Prompt {
	return &Prompt{
		queue: strings.Builder{},
	}
}

/*
AddRoleTemplate adds a role-specific template to the prompt.
It retrieves the template from the configuration based on the given role
and appends it to the prompt queue.
*/
func (p *Prompt) AddRoleTemplate(role RoleType) *Prompt {
	template := viper.GetString(fmt.Sprintf("prompt.template.role.%s", getRoleString(role)))
	p.queue.WriteString(template)
	p.queue.WriteString("\n\n")
	return p
}

/*
AddScratchpad adds a scratchpad template to the prompt with the given context.
It retrieves the scratchpad template from the configuration and replaces the
{context} placeholder with the provided context.
*/
func (p *Prompt) AddScratchpad(context string) *Prompt {
	template := viper.GetString("prompt.template.scratchpad")
	p.queue.WriteString(strings.Replace(template, "{context}", context, 1))
	p.queue.WriteString("\n\n")
	return p
}

/*
AddContent adds content-specific template to the prompt.
It retrieves the template for the given content type from the configuration
and replaces the {contentType} placeholder with the provided content.
*/
func (p *Prompt) AddContent(contentType string, content string) *Prompt {
	template := viper.GetString(fmt.Sprintf("prompt.template.content.%s", contentType))
	p.queue.WriteString(strings.Replace(template, fmt.Sprintf("{%s}", contentType), content, 1))
	p.queue.WriteString("\n\n")
	return p
}

/*
AddInstructions adds the instructions template to the prompt.
It retrieves the instructions template from the configuration and appends
it to the prompt queue.
*/
func (p *Prompt) AddInstructions() *Prompt {
	instructions := viper.GetString("prompt.template.instructions")
	p.queue.WriteString(instructions)
	p.queue.WriteString("\n\n")
	return p
}

/*
Build constructs and returns the final prompt string.
*/
func (p *Prompt) Build() string {
	return p.queue.String()
}

/*
getRoleString converts a RoleType to its corresponding string representation.
*/
func getRoleString(role RoleType) string {
	switch role {
	case CODER:
		return "coder"
	case REVIEWER:
		return "reviewer"
	default:
		return "unknown"
	}
}
