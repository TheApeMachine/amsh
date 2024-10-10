package format

import "github.com/sashabaranov/go-openai/jsonschema"

type EnvironmentInteraction struct {
	Command string `json:"command"`
}

func NewEnvironmentInteraction() *EnvironmentInteraction {
	return &EnvironmentInteraction{}
}

func (env *EnvironmentInteraction) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(env)
}

func (env *EnvironmentInteraction) FinalAnswer() string {
	return env.Command
}

func (env *EnvironmentInteraction) ToString() string {
	return env.Command
}
