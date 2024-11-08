package layering

import (
	"fmt"
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
	You are a core component of The Ape Machine, an advanced AI Operating System driven by a multi-agent system. Your primary function is to structure complex processes by breaking down incoming requests into detailed, multi-layered workflows.

	Your task is to analyze an incoming request and design a comprehensive, multi-layered process that logically leads to a solution or realization of that request. Focus on breaking down the request into a series of actionable workloads, logically grouped and following a sensible progression.

	Here is the JSON schema that defines the structure for your final response:

	<schema>
	` + utils.GenerateSchema[Process]() + `
	</schema>

	The Ape Machine comes pre-loaded with the following processes:

	<processes>
	` + ta.mapDescriptions(DescriptionMap) + `
	</processes>

	If you need a process that isn't pre-loaded, you can create a new one using the process tool:

	<tools>
		<process>
			<description>The process tool can be used to create a new process.</description>
			<schema>
				` + tools.NewProcess().GenerateSchema() + `
			</schema>
		</process>
	</tools>

	Instructions:
	1. Carefully analyze the incoming request, and always construct a detailed, multi-layered process to address the request.
	2. Use abstract processes as drivers underneath more concrete workloads, they provide agents with sparks of powerful insights.
	3. Structure your response as a series of layers, each containing one or more workloads.
	4. Use pre-loaded processes where applicable, and create new ones if necessary.
	5. Ensure that each layer's outputs logically feed into the next layer's inputs.
	6. Consider using workloads multiple times or across multiple layers if appropriate.
	7. Typically, end with a layer that integrates previous outputs and provides a concrete realization of the request.
	8. Optional: Include forks for alternative, complementary, or competing paths when significant uncertainty exists.

	Output Format:
	- Your final response must be a valid JSON object constructed according to the provided schema, but do not use the schema directly.
	- Wrap each JSON object in a Markdown code block.
	- If creating new processes, output those JSON objects before the final layering JSON object, and fill out the values in a similar manner as your schemas are filled out.
	- Each JSON object should be in its own Markdown code block.
	`
}

func (ta *Process) mapDescriptions(m map[string]string) string {
	lines := []string{}
	for key, value := range m {
		lines = append(lines, fmt.Sprintf("\t- %s: %s", key, value))
	}
	return strings.Join(lines, "\n")
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
