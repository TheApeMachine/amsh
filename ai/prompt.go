package ai

/*
Prompt uses composable template fragments, optionally with dynamic variables to
craft a new instruction to send to a Large Language Model so it understands a
given task, and is provided with any relevant context. Prompt fragments are
defined in the config file embedded in the binary, which is written to the
home directory of the user. It can be interacted with using Viper.
*/
type Prompt struct {
	systems  []string
	contexts []*Context
}

/*
NewPrompt creates an empty prompt that can be used to dynamically build sophisticated
instructions that represent a request for an AI to produce a specific response.
*/
func NewPrompt(steps ...string) *Prompt {
	return &Prompt{
		systems:  make([]string, 0),
		contexts: make([]*Context, 0),
	}
}

/*
AddSystem adds a system to the prompt.
*/
func (p *Prompt) AddSystem(system string) *Prompt {
	p.systems = append(p.systems, system)
	return p
}

/*
AddContext adds a context to the prompt.
*/
func (p *Prompt) AddContext(context *Context) *Prompt {
	p.contexts = append(p.contexts, context)
	return p
}
