package process

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

type Moonshot struct {
	Ideas []string `json:"ideas" jsonschema:"required;title=Moonshot Ideas;description=A list of moonshot ideas."`
}

func NewMoonshot() *Moonshot {
	return &Moonshot{}
}

func (moonshot *Moonshot) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.moonshot.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", moonshot.GenerateSchema())
	return prompt
}

func (moonshot *Moonshot) GenerateSchema() string {
	schema := jsonschema.Reflect(&Moonshot{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type Sensible struct {
	Ideas []string `json:"ideas" jsonschema:"required;title=Sensible Ideas;description=A list of sensible ideas."`
}

func NewSensible() *Sensible {
	return &Sensible{}
}

func (sensible *Sensible) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.sensible.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", sensible.GenerateSchema())
	return prompt
}

func (sensible *Sensible) GenerateSchema() string {
	schema := jsonschema.Reflect(&Sensible{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type Catalyst struct {
	Ideas []string `json:"ideas" jsonschema:"required;title=Catalyst Ideas;description=A list of catalyst ideas."`
}

func NewCatalyst() *Catalyst {
	return &Catalyst{}
}

func (catalyst *Catalyst) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.catalyst.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", catalyst.GenerateSchema())
	return prompt
}

func (catalyst *Catalyst) GenerateSchema() string {
	schema := jsonschema.Reflect(&Catalyst{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type Guardian struct {
	Ideas []string `json:"ideas" jsonschema:"required;title=Guardian Ideas;description=A list of guardian ideas."`
}

func NewGuardian() *Guardian {
	return &Guardian{}
}

func (guardian *Guardian) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.guardian.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", guardian.GenerateSchema())
	return prompt
}

func (guardian *Guardian) GenerateSchema() string {
	schema := jsonschema.Reflect(&Guardian{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type Ideas struct {
	Pitch      string     `json:"pitch" jsonschema:"required;title=Pitch;description=A short pitch for the ideas."`
	Motivation string     `json:"motivation" jsonschema:"required;title=Motivation;description=The motivation for the ideas."`
	Arguments  []Argument `json:"arguments" jsonschema:"required;title=Arguments;description=The arguments for the ideas."`
}

type Argument struct {
	Reason   string `json:"reason" jsonschema:"required;title=Reason;description=The reason for the argument."`
	Evidence string `json:"evidence" jsonschema:"required;title=Evidence;description=The evidence for the argument."`
}
