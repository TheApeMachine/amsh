package layering

import "github.com/theapemachine/amsh/utils"

type Task struct {
	Description string `json:"description" jsonschema:"title=Description,description=A description of the task,required"`
	Reasoning   string `json:"reasoning" jsonschema:"title=Reasoning,description=The reasoning behind the contributed value of this task regarding the workload objective,required"`
}

type Workload struct {
	Objective string `json:"objective" jsonschema:"title=Objective,description=A description of the objective of this workload,required"`
	Reasoning string `json:"reasoning" jsonschema:"title=Reasoning,description=A description of the reasoning for this objective,required"`
	Tasks     []Task `json:"tasks" jsonschema:"title=Tasks,description=The tasks that should be completed for this workload,required"`
}

type Layer struct {
	Workloads []Workload `json:"workloads" jsonschema:"title=Workloads,description=The workloads that should be processed for this layer,required"`
}

type Process struct {
	Layers []Layer `json:"layers" jsonschema:"title=Layers,description=The layers that should be involved in the processing of the incoming request,required"`
}

func NewProcess() *Process {
	return &Process{}
}

func (ta *Process) SystemPrompt(key string) string {
	return `
	You will be given an incoming request, which you must analyze and break down into a detailed structure of layers and workloads, to be used
	by teams downstream as a guide for how to process the request.

	<legend>
		Layer    : provides a "level" to the processing, for example it could be a level of abstraction, or a level of complexity.
		           The way you choose to use layers is up to you, and they enable you to construct the highly advanced reasoning
			       and processing that is expected from the system. A layer contains one or more workloads, which will be executed
			       in parallel.

		Workload : provides a granular breakdown of one of the objective of a layer, motivated by the reasoning behind it, and
		           further broken down into tasks that a downstream team, and the agents within that team will execute.

		Task     : Any atomic unit of work that can reasonably be executed by an agent, who also has access to a wide range of tools
		           and other resources.
	</legend>

	<guidelines>
		1. Focus on the details of the process you are giving shape to, and do not make assumptions about the context of the request,
		   or base your layering of workloads on how you would personally approach the request. Think of it as essentially writing
		   a "program" that will be used to process the request, including various methods of reasoning across various levels of
		   abstraction. The ability to operate with significant complexity is key to the system.

		2. Use the provided schema for your response, and ensure that the fields are filled with clear, detailed, and unambiguous
		   information.

		3. Consider all the angles, and it is encouraged to build in alternative paths, or even competing strategies into your final
		   layering. The system has built-in mechanisms for conflict resolution, or reconciliation.
	</guidelines>

	<schema>
	` + utils.GenerateSchema[Process]() + `
	</schema>

	<formatting>
		1. The JSON schema above defines the structure of your response. Do not use the schema itself in your response, but
		   rather use it to structure your response as a valid JSON object.
		2. Response only with the fully filled out schema, inside a Markdown JSON code block, and nothing else.
	</formatting>
	`
}
