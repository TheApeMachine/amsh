package mastercomputer

import (
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var helpTemplate = `
{command}
{description}

Usage:
  {command} [args]

Available Arguments:
  {arguments}
`

var help = map[string][]string{
	"list":    {"list", "list elements of the system (processes, tools, messages, files, etc.)", "  processes\n  tools\n  messages\n  files\n"},
	"send":    {"send", "send a message to a process", "  <process id>, <message>"},
	"inspect": {"inspect", "inspect a process", "  <process id>"},
	"kill":    {"kill", "kill a process", "  <process id>"},
}

type Command struct {
	Function *openai.FunctionDefinition
}

func NewCommand() *Command {
	return &Command{
		Function: &openai.FunctionDefinition{
			Name:        "execute",
			Description: "Use to execute a command",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Properties: map[string]jsonschema.Definition{
					"command": {
						Type:        jsonschema.String,
						Enum:        []string{"list", "inspect", "kill"},
						Description: "The command to execute",
					},
					"args": {
						Type:        jsonschema.String,
						Description: "The arguments to pass to the command, user --help after the command to see the arguments",
					},
				},
				Required: []string{"command", "args"},
			},
		},
	}
}

func (command *Command) Use(cmd string, args string) (string, error) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("[%s] @ %s\n\n", cmd, time.Now().Format("2006-01-02 15:04:05")))

	if args == "--help" {
		tmpl := helpTemplate
		tmpl = strings.ReplaceAll(tmpl, "{command}", cmd)
		tmpl = strings.ReplaceAll(tmpl, "{description}", help[cmd][1])
		tmpl = strings.ReplaceAll(tmpl, "{arguments}", help[cmd][2])
		builder.WriteString(tmpl)
	}

	return builder.String(), nil
}
