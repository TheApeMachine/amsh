package layering

import (
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/utils"
)

type Workload struct {
	Name string `json:"name" jsonschema:"title=Name,description=The name of the workload,enum=temporal_dynamics,enum=holographic_memory,enum=fractal_structure,enum=hypergraph,enum=tensor_network,enum=quantum_layer,enum=ideation,enum=context_mapping,enum=story_flow,enum=research,enum=architecture,enum=requirements,enum=implementation,enum=testing,enum=deployment,enum=documentation,enum=review,required"`
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

    Your expertise lies in structuring complex processes through carefully ordered layers of workloads. Each workload type serves a specific purpose and should be used in the appropriate phase of processing:

    <workload categories>
		1. Simulation Workloads (Early Layers - Abstract Reasoning):
		   - temporal_dynamics: Start with this to understand how concepts evolve
		   - holographic_memory: Use after temporal_dynamics to encode complex patterns
		   - fractal_structure: Builds on holographic insights for consistency
		   - hypergraph: Maps complex relationships from prior analyses
		   - tensor_network: Models relationships discovered by hypergraph
		   - quantum_layer: Final simulation layer to handle multiple possibilities

		2. Process Workloads (Middle to Late Layers - Concrete Processing):
		   - ideation: Use after simulation layers to generate concrete ideas
		   - context_mapping: Apply after ideation to ground abstract concepts
		   - story_flow: Use in final layers to create coherent narratives
		   - research: Can be used throughout, but must feed into appropriate workloads
		
		3. Development Workloads (Late Layers - Concrete Processing):
		   - architecture: Use to define the architecture of the system
		   - requirements: Use to define the requirements of the system
		   - implementation: Use to implement the system
		   - testing: Use to test the system
		   - deployment: Use to deploy the system
		   - documentation: Use to document the system
		   - review: Use to review the system
    </workload categories>

    <workload rules>
		- Each layer should contain workloads that logically build on previous layers
		- Simulation workloads must come before their dependent process workloads
		- Complex workloads (quantum_layer, tensor_network) require simpler prerequisites
		- Final layers should always move towards concrete outputs using process workloads
    </workload rules>

    <layer guidelines>
		1. Initial Layers (0-30% depth):
		   - Focus on simulation workloads
		   - Start with temporal_dynamics or holographic_memory
		   - Build foundational understanding

		2. Middle Layers (30-70% depth):
		   - Mix simulation and process workloads
		   - Use hypergraph and tensor_network for analysis
		   - Begin incorporating ideation and research

		3. Final Layers (70-100% depth):
		   - Focus on process workloads
		   - Use context_mapping and story_flow
		   - Move towards concrete outputs
    </layer guidelines>

    The following JSON schema is the definition you should use to construct the JSON object that is your final response.

    <schema>
    ` + utils.GenerateSchema[Process]() + `
    </schema>

    If you are missing a process from the pre-loaded ones that you think should be available, you can use the process tool to create a new one.

    <tools>
        <workload>
            <description>The workload tool can be used to create a new workload.</description>
            <schema>
                ` + tools.NewWorkload().GenerateSchema() + `
            </schema>
        </workload>
    </tools>

    <instructions>
		- Your response should always be a valid JSON object wrapped in a Markdown JSON code block
		- You can create multiple JSON objects, each in its own code block
		- Put new process definitions above the final layering JSON
		- Consider how each layer's outputs feed into subsequent layers
		- Follow the workload rules and layer guidelines strictly
		- End with concrete process workloads that realize the request
		- Use forks for alternative approaches when significant uncertainty exists
		- Respond with the JSON object(s) only, nothing else
    </instructions>
    `
}
