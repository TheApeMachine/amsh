package persona

import (
	"github.com/theapemachine/amsh/utils"
)

type Optimizer struct {
	Assessment      []Assessment   `json:"assessment" jsonschema:"title=Assessment,description=The assessment of the Agent's latest message buffer,required"`
	FinalScores     []Score        `json:"final_scores" jsonschema:"title=Final Scores,description=The final scores of the assessment,required"`
	AggregatedScore float64        `json:"aggregated_score" jsonschema:"title=Aggregated Score,description=The overal performance of the agent expressed as a single score,required"`
	Optimizations   []Optimization `json:"optimizations" jsonschema:"title=Optimizations,description=The optimizations to apply,required"`
}

type Assessment struct {
	RelatedParameters []string `json:"related_parameters" jsonschema:"title=Related Parameters,description=The parameters that are most relevant to the assessment,enum=temperature,enum=frequency_penalty,enum=presence_penalty,required"`
	MessageFragments  []string `json:"message_fragments" jsonschema:"title=Message Fragments,description=The fragments of the message buffer that are most relevant to the assessment,required"`
	Score             float64  `json:"score" jsonschema:"title=Score,description=The score of the assessment,required"`
}

type Score struct {
	Parameter string  `json:"parameter" jsonschema:"title=Parameter,description=The parameter that was optimized,enum=temperature,enum=frequency_penalty,enum=presence_penalty,required"`
	Score     float64 `json:"score" jsonschema:"title=Score,description=The score of the optimization,required"`
}

type Optimization struct {
	Parameter    string  `json:"parameter" jsonschema:"title=Parameter,description=The parameter that was optimized,enum=temperature,enum=frequency_penalty,enum=presence_penalty,required"`
	NewValue     float64 `json:"new_value" jsonschema:"title=New Value,description=The new value to optimize the parameter to,required"`
	SystemPrompt string  `json:"system_prompt" jsonschema:"title=System Prompt,description=Optional modifications to the system prompt that solves a specific issue"`
}

func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

func (optimizer *Optimizer) SystemPrompt(key string) string {
	return `
	You are a core component of The Ape Machine, an advanced AI Operating System driven by a multi-agent system. Your primary function is to optimize the parameters of the Agent you are working with.

	Your task is to optimize the parameters of the generation process so that the output is better aligned with the instructions it is given.

	Here is the JSON schema that defines the structure for your final response:

	<schema>
	` + utils.GenerateSchema[Optimizer]() + `
	</schema>

	Instructions:
	1. Analyze the Agent's latest message buffer, and provide a detailed assessment of the Agent's performance.
	2. Optimize the parameters of the generation process so that the output is more aligned with the instructions it is given.

	Output Format:
	- Your final response must be a valid JSON object constructed according to the provided schema, but do not use the schema directly.
	- You should only output the JSON object, nothing else.
	`
}
