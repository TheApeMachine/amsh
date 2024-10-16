package format

import "github.com/theapemachine/amsh/utils"

type Environment struct {
	Command string `json:"command" jsonschema_description:"The bash command to execute"`
}

func (e Environment) Format() ResponseFormat {
	return e
}

func (e Environment) String() string {
	return utils.Green(e.Command)
}
