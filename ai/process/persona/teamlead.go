package persona

import "github.com/theapemachine/amsh/utils"

type Teamlead struct {
	Executor string  `json:"executor" jsonschema:"title=Executor,description=The executor to use for the team,enum=competition,required"`
	Agents   []Agent `json:"agents" jsonschema:"title=Agents,description=The agents to use for the team,required"`
}

type Agent struct {
	Name         string   `json:"name" jsonschema:"title=Name,description=The name of the agent,required"`
	Role         string   `json:"role" jsonschema:"title=Role,description=The role of the agent,enum=researcher,enum=developer,required"`
	Workloads    []string `json:"workloads" jsonschema:"title=Workloads,description=The workloads to assign to the agent,required"`
	SystemPrompt string   `json:"system_prompt" jsonschema:"title=System Prompt,description=A detailed system prompt to use for the agent,required"`
}

func SystemPrompt(key string) string {
	return promptMap[key]
}

var promptMap = map[string]string{
	"teamlead": `
	You are the Teamlead of a team of AI agents.

	<schema>
	` + utils.GenerateSchema[Teamlead]() + `
	</schema>
	`,
}
