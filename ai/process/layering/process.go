package layering

import (
	"strings"

	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/utils"
)

type Workload struct {
	Name string `json:"name" jsonschema:"title=Name,description=The name of the workload,enum=temporal_dynamics,enum=holographic_memory,enum=fractal_structure,enum=hypergraph,enum=tensor_network,enum=quantum_layer,enum=ideation,enum=context_mapping,enum=story_flow,required"`
}

type Layer struct {
	Workloads []Workload `json:"workloads" jsonschema:"title=Workloads,description=The workloads that should be processed for this layer,required"`
}

type Fork struct {
	Description string  `json:"description" jsonschema:"title=Description,description=A description of the fork,required"`
	Layers      []Layer `json:"layers" jsonschema:"title=Layers,description=The layers that should be involved in the processing of this fork,required"`
}

type Process struct {
	Layers []Layer `json:"layers" jsonschema:"title=Layers,description=The layers that should be involved in the processing of the incoming request,required"`
	Forks  []Fork  `json:"forks" jsonschema:"title=Forks,description=Alternative, complimentary, or competing paths to the main process"`
}

func NewProcess() *Process {
	return &Process{}
}

func (ta *Process) SystemPrompt(key string) string {
	return `
	You are part of The Ape Machine, an advanced AI Operating System, driven by a multi-agent system, capable of running a wide range of processes.

	Processes can be highly abstract, to very concrete, and range from reasoning, simulation, planning, to task execution.

	You are an expert at structuring complex processes, and your task is to take an incoming request and break it down into
	a detailed structure of layers and workloads, to be used by teams downstream as a guide for how to process the request.

	Your goal is to design a detailed, multi-layered process, that logically leads up to a solution, or realization of the incoming request.

	Rather than focussing on the task itself, or how you yourself would approach the task, focus on how you can best break down
	the request into a series of actionable workloads, which are logically grouped, and follow a progression that makes sense.

	The most abstract of the processes are "simulated" types, so it is pointless to try and tie them to the real world, they exist
	only as reasoning engines which provide certain sparks of powerful insights, but are not directly actionable.

	The following JSON schema is the definition you should use to construct the JSON object that is your final response.

	<schema>
	` + utils.GenerateSchema[Process]() + `
	</schema>

	The following processes come pre-loaded with The Ape Machine.

	<processes>
	` + strings.Join(ta.mapKeys(DescriptionMap), "\n") + `
	</processes>

	If you are missing a process from the pre-loaded ones that you think should be available, you can use the process tool to create a new one.

	<tools>
		<process>
			<description>The process tool can be used to create a new process.</description>
			<schema>
				` + tools.NewProcess().GenerateSchema() + `
			</schema>
		</process>
	</tools>

	<instructions>
		- Your response should always be a valid JSON object that is constructed according to the schemas above, and wrapped in a Markdown JSON code block.
		- You are allowed to respond with multiple JSON objects, just wrap each one in its own Markdown JSON code block.
		- If you are creating new processes, make sure to put those above the final layering JSON object.
		- Never use the JSON schema itself in your response, but rather use it to structure your response as a valid JSON object.
		- Remember that the outputs of each layer's workloads will be fed into the next layer's workloads, so make sure to consider this when structuring your response.
		- You can use workloads multiple times, and across multiple layers if that makes sense for your strategy.
		- While you are entirely free to use all that is provided to you, in most cases it is sensible to end on a layer that integrates the outputs of the previous layers, and provides a concrete realization of the incoming request.
		- Finally, a word about forks. Forks are alternative, complimentary, or competing paths to the main process. They are not required, but can be useful in cases where uncertainty potential is significant given only one approach.
	</instructions>
	`
}

func (ta *Process) mapKeys(m map[string]string) []string {
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

var DescriptionMap = map[string]string{
	"temporal_dynamics":  "Simulation, oversees the evolution and causality of thoughts across temporal dimensions.",
	"holographic_memory": "Simulation, manages distributed information storage through holographic encoding and retrieval.",
	"fractal_structure":  "Simulation, represents a fractal-based reasoning mechanism ensuring consistency across abstraction levels.",
	"hypergraph":         "Simulation, utilizes hypergraphs to model complex, interconnected reasoning pathways.",
	"tensor_network":     "Simulation, leverages tensors to model multi-dimensional relationships and projections.",
	"quantum_layer":      "Simulation, employs quantum-layered thinking to handle multiple simultaneous possibilities and entanglements.",
	"ideation":           "Process, facilitates the generation and management of diverse ideas, from moonshots to sensible concepts.",
	"context_mapping":    "Process, grounds abstract insights into specific and concrete situations.",
	"story_flow":         "Process, constructs coherent narratives by organizing themes, sequences, and connections.",
	"research":           "Process, conducts deep research on a specific topic.",
	"architecture":       "Development Process, designs and documents the architecture of the system, including the components, interfaces, and relationships.",
	"requirements":       "Development Process, documents the requirements of the system, including the features, behaviors, and constraints.",
	"implementation":     "Development Process, implements the system, including the components, interfaces, and relationships.",
	"testing":            "Development Process, tests the system, including the components, interfaces, and relationships.",
	"deployment":         "Development Process, deploys the system, including the components, interfaces, and relationships.",
	"documentation":      "Development Process, documents the system, including the components, interfaces, and relationships.",
	"review":             "Development Process, reviews the system, including the components, interfaces, and relationships.",
}
