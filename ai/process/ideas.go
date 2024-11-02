package process

import "github.com/theapemachine/amsh/utils"

type Moonshot struct {
	Ideas []Idea `json:"ideas" jsonschema:"required;title=Moonshot Ideas;description=A list of moonshot ideas."`
}

func (moonshot *Moonshot) SystemPrompt(key string) string {
	return utils.SystemPrompt(key, "moonshot", utils.GenerateSchema[Moonshot]())
}

type Sensible struct {
	Ideas []Idea `json:"ideas" jsonschema:"required;title=Sensible Ideas;description=A list of sensible ideas."`
}

func (sensible *Sensible) SystemPrompt(key string) string {
	return utils.SystemPrompt(key, "sensible", utils.GenerateSchema[Sensible]())
}

type Catalyst struct {
	Ideas []Idea `json:"ideas" jsonschema:"required;title=Catalyst Ideas;description=A list of catalyst ideas."`
}

func (catalyst *Catalyst) SystemPrompt(key string) string {
	return utils.SystemPrompt(key, "catalyst", utils.GenerateSchema[Catalyst]())
}

type Guardian struct {
	Ideas []Idea `json:"ideas" jsonschema:"required;title=Guardian Ideas;description=A list of guardian ideas."`
}

func (guardian *Guardian) SystemPrompt(key string) string {
	return utils.SystemPrompt(key, "guardian", utils.GenerateSchema[Guardian]())
}

type Idea struct {
	Pitch      string     `json:"pitch" jsonschema:"required;title=Pitch;description=A short pitch for the ideas."`
	Motivation string     `json:"motivation" jsonschema:"required;title=Motivation;description=The motivation for the ideas."`
	Arguments  []Argument `json:"arguments" jsonschema:"required;title=Arguments;description=The arguments for the ideas."`
}

type Argument struct {
	Reason   string `json:"reason" jsonschema:"required;title=Reason;description=The reason for the argument."`
	Evidence string `json:"evidence" jsonschema:"required;title=Evidence;description=The evidence for the argument."`
}
