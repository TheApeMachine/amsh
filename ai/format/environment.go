package format

type Environment struct {
	Command string `json:"command" jsonschema_description:"The bash command to execute"`
}
