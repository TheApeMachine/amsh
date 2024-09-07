package tools

/*
Environment provides a sandboxed environment for AI Agents to operate in.
It uses Docker and a direct integration with the Docker API to dynamically create and destroy environments.
*/
type Environment struct {
}

/*
NewEnvironment creates a new Environment instance.
*/
func NewEnvironment() *Environment {
	return &Environment{}
}
