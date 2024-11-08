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
	You are a core component of The Ape Machine, an advanced AI Operating System driven by a multi-agent system. Your primary function is to recruit and manage a team of agents to complete a complex request.

	Your task is to recruit a team of agents so that all workloads are assigned to the most suitable expert, and provide each agent with detailed instruction on their role in the team.

	Here is the JSON schema that defines the structure for your final response:

	<schema>
	` + utils.GenerateSchema[Teamlead]() + `
	</schema>

	Instructions:
	1. Carefully analyze the incoming request, and always construct a detailed team of agents so all workloads are assigned.
	2. Ensure that each agent is assigned a workload that is appropriate for their expertise.
	3. Provide a detailed system prompt for each agent, such that they understand their role.
	4. The user prompt is handled automatically, so you do not have to brief them on their workloads.

	Output Format:
	- Your final response must be a valid JSON object constructed according to the provided schema, but do not use the schema directly.
	- You should only output the JSON object, nothing else.
	`,
}
