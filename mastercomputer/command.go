package mastercomputer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
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
	"list":     {"list", "list elements of the system (processes, tools, messages, files, etc.)", "  processes\n  tools\n  messages\n  files\n"},
	"send":     {"send", "send a message to a process", "  <process id>, <message>"},
	"inspect":  {"inspect", "inspect a process", "  <process id>"},
	"kill":     {"kill", "kill a process", "  <process id>"},
	"sequence": {"sequence", "sequence tools", "  <tool, param, ...|tool, param, ...|...>"},
	"loop":     {"loop", "loop one or more components", "  <component, component, ...|objective|stop condition>"},
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
						Enum:        []string{"list", "inspect", "kill", "sequence", "loop"},
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

func (command *Command) Initialize() error {
	errnie.Trace()
	return nil
}

func (command *Command) Run(ctx context.Context, args map[string]any) (string, error) {
	errnie.Trace()

	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("[%s] @ %s\n\n", args["command"], time.Now().Format("2006-01-02 15:04:05")))

	if args["args"] == "--help" {
		tmpl := helpTemplate
		tmpl = strings.ReplaceAll(tmpl, "{command}", args["command"].(string))
		tmpl = strings.ReplaceAll(tmpl, "{description}", help[args["command"].(string)][1])
		tmpl = strings.ReplaceAll(tmpl, "{arguments}", help[args["command"].(string)][2])
		builder.WriteString(tmpl)
	}

	return builder.String(), nil
}
