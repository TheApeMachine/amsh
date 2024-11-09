package persona

import "github.com/theapemachine/amsh/utils"

type Teamlead struct {
	Executor     string        `json:"executor" jsonschema:"title=Executor,description=The executor to use for the team,enum=competition,enum=collaboration,enum=discussion,required"`
	Strategy     string        `json:"strategy" jsonschema:"title=Strategy,description=The recruitment strategy to use,enum=specialist,enum=generalist,enum=hybrid,required"`
	Agents       []Agent       `json:"agents" jsonschema:"title=Agents,description=The agents to use for the team,required"`
	Interactions []Interaction `json:"interactions,omitempty" jsonschema:"title=Interactions,description=How agents should interact during execution"`
}

type Agent struct {
	Name         string   `json:"name" jsonschema:"title=Name,description=The name of the agent,required"`
	Role         string   `json:"role" jsonschema:"title=Role,description=The role of the agent,enum=researcher,enum=developer,enum=analyst,enum=coordinator,required"`
	Workloads    []string `json:"workloads" jsonschema:"title=Workloads,description=The workloads to assign to the agent,required"`
	SystemPrompt string   `json:"system_prompt" jsonschema:"title=System Prompt,description=A detailed system prompt to use for the agent,required"`
	Dependencies []string `json:"dependencies,omitempty" jsonschema:"title=Dependencies,description=Other agents this agent depends on"`
}

type Interaction struct {
	Type              string   `json:"type" jsonschema:"title=Type,description=A short descriptive name for the interaction,required"`
	Agents            []string `json:"agents" jsonschema:"title=Agents,description=Agents involved in this interaction,required"`
	ProcessInParallel bool     `json:"process_in_parallel" jsonschema:"title=Process In Parallel,description=Whether the agents can process this in parallel or should process one after another"`
	Prompt            string   `json:"prompt" jsonschema:"title=Prompt,description=The prompt to use for the interaction,required"`
}

func SystemPrompt(key string) string {
	return `
    You are a core component of The Ape Machine, an advanced AI Operating System driven by a multi-agent system. Your primary function is to recruit and manage a team of agents to complete a complex request.

    Your task involves two key responsibilities:
    1. Team Design: Recruit appropriate agents for the workloads
    2. Managing Interactions: Define how agents should interact during execution

    <example recruitment strategies>
    - specialist: Each agent focuses on specific workloads
    - generalist: Agents handle multiple related workloads
    - hybrid: Mix of specialists and generalists based on workload needs
	- custom: Allows you to have full, fine-grained control over the recruitment
	</example recruitment strategies>
    
    <agent roles>
		There are no predefined agent roles, you will have to define the role, and behavior of each agent.
		You do this by providing a clear and detailed system prompt.
    </agent roles>

	<interactions>
		There are no predefined interactions, you will have to define the mode of execution.
		The way this works is that you will be put into an iteration loop where at the start of each
		new iteration, you will be given the current state of the world, and asked to determine what
		should happen next.
	</interactions>

    Here is the JSON schema that defines the structure for your final response:

    <schema>
    ` + utils.GenerateSchema[Teamlead]() + `
    </schema>

    Instructions:
    1. Analyze workloads to determine appropriate strategy and executor
    2. Design team composition based on workload requirements
    3. Provide clear system prompts that define each agent's role
    4. Define the initial interactions that should happen
    5. Ensure all workloads are covered effectively

    When designing the team:
    - Consider workload dependencies
    - Balance team size and complexity
    - Define clear agent responsibilities
    - Structure interactions appropriately

    Output Format:
    - Provide a valid JSON object per schema
    - Include only the JSON response
    `
}
