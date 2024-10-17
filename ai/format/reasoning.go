package format

import (
	"github.com/theapemachine/amsh/utils"
)

type Reasoning struct {
	Strategies []Strategy `json:"strategies" jsonschema:"title=Strategies,description=The strategies to be used"`
	Done       bool       `json:"done" jsonschema:"title=Done,description=Whether the reasoning is done"`
}

func (r Reasoning) String() string {
	output := utils.Muted("[REASONING]\n")

	for _, strategy := range r.Strategies {
		for _, thought := range strategy.Thoughts {
			output += utils.Blue(thought.Thought + "\n")

			output += utils.Red("  Reflection:\n")
			output += utils.Yellow("    Thought: " + thought.Reflection.Thought + "\n")
			output += utils.Green("    Verify: " + thought.Reflection.Verify + "\n")

			output += utils.Red("  Challenge:\n")
			output += utils.Yellow("    Thought: " + thought.Challenge.Thought + "\n")
			output += utils.Green("    Debate:\n")
			for _, argument := range thought.Challenge.Debate {
				output += utils.Red("      Argument: " + argument.Argument + "\n")
				output += utils.Green("      Counter: " + argument.Counter + "\n")
			}
			output += utils.Blue("    Resolve: " + thought.Challenge.Resolve + "\n")

			output += "\n"
		}
	}

	output += utils.Muted("[/REASONING]\n")
	return output
}

type Strategy struct {
	Thoughts []Thought `json:"thoughts" jsonschema:"title=Thoughts,description=The thoughts you have while using the strategy"`
}

type Thought struct {
	Thought    string     `json:"thought" jsonschema:"title=Thought,description=A single thought"`
	Reflection Reflection `json:"reflection" jsonschema:"title=Reflection,description=The reflection of the thought"`
	Challenge  Challenge  `json:"challenge" jsonschema:"title=Challenge,description=The challenge of the thought"`
}

type Reflection struct {
	Thought string `json:"thought" jsonschema:"title=Thought,description=Restate the thought in a more clear and concise way"`
	Verify  string `json:"verify" jsonschema:"title=Verify,description=Verify the thought is correct"`
}

type Challenge struct {
	Thought string     `json:"thought" jsonschema:"title=Thought,description=Restate the thought in a more clear and concise way"`
	Debate  []Argument `json:"debate" jsonschema:"title=Debate,description=Debate the thought with an external voice"`
	Resolve string     `json:"resolve" jsonschema:"title=Resolve,description=Resolve the thought"`
}

type Argument struct {
	Argument string `json:"argument" jsonschema:"title=Argument,description=An argument for the thought"`
	Counter  string `json:"counter" jsonschema:"title=Counter,description=A counter argument to the thought"`
}
