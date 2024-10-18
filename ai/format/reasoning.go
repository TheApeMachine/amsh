package format

import (
	"github.com/theapemachine/amsh/utils"
)

type Reasoning struct {
	Strategies  []Strategy `json:"strategies" jsonschema:"description=A dynamically constructed strategy to solve the problem"`
	FinalAnswer string     `json:"final_answer" jsonschema:"description=The final answer to the question"`
	Done        bool       `json:"done" jsonschema:"description=You will have infinite iterations to reason, until you set this to true"`
	NextSteps   []Step     `json:"next_steps" jsonschema:"description=The next steps to be taken"`
}

func (r Reasoning) String() string {
	output := utils.Dark("[REASONING]") + "\n"

	for _, strategy := range r.Strategies {
		for _, thought := range strategy.Thoughts {
			output += "\t" + utils.Red(thought.Thought) + "\n"

			output += "\t" + utils.Muted("[REFLECTION]") + "\n"
			output += "\t\t" + utils.Yellow("Thought: ") + thought.Reflection.Thought + "\n"
			output += "\t\t" + utils.Green("Verify : ") + thought.Reflection.Verify + "\n"
			output += "\t" + utils.Muted("[/REFLECTION]") + "\n"

			output += "\t" + utils.Muted("[CHALLENGE]") + "\n"
			output += "\t\t" + utils.Red("Thought: ") + thought.Challenge.Thought + "\n"
			output += "\t\t" + utils.Muted("[DEBATE]") + "\n"
			for _, argument := range thought.Challenge.Debate {
				output += "\t\t\t" + utils.Yellow("Argument: ") + argument.Argument + "\n"
				output += "\t\t\t" + utils.Green("Counter : ") + argument.Counter + "\n"
			}
			output += "\t\t" + utils.Muted("[/DEBATE]") + "\n"
			output += "\t\t" + utils.Blue("Resolve: ") + thought.Challenge.Resolve + "\n"
			output += "\t" + utils.Muted("[/CHALLENGE]") + "\n"
		}
	}

	output += "\t" + utils.Blue("Final Answer: ") + r.FinalAnswer + "\n"
	output += "\t" + utils.Red("Done: ") + BoolToString(r.Done) + "\n"
	output += utils.Dark("[/REASONING]") + "\n"
	return output
}

type Strategy struct {
	Thoughts []Thought `json:"thoughts" jsonschema:"description=The thoughts you have while using the strategy"`
}

type Thought struct {
	Thought     string     `json:"thought" jsonschema:"description=A single thought"`
	Reflection  Reflection `json:"reflection" jsonschema:"description=The reflection of the thought"`
	OutOfTheBox string     `json:"out_of_the_box" jsonschema:"description=An out of the box idea"`
	Challenge   Challenge  `json:"challenge" jsonschema:"description=The challenge of the thought"`
}

type Reflection struct {
	Thought string `json:"thought" jsonschema:"description=Restate the thought in a more clear and concise way"`
	Verify  string `json:"verify" jsonschema:"description=Verify the thought is correct"`
}

type Challenge struct {
	Thought string   `json:"thought" jsonschema:"description=Restate the thought from a different perspective"`
	Debate  []Debate `json:"debate" jsonschema:"description=Debate the thought with an external voice and multiple arguments"`
	Resolve string   `json:"resolve" jsonschema:"description=Resolve the debate from an independent perspective"`
}

type Debate struct {
	Argument string `json:"argument" jsonschema:"description=An argument for the thought"`
	Counter  string `json:"counter" jsonschema:"description=A counter argument to the thought"`
	Answer   string `json:"answer" jsonschema:"description=An answer considering the counter argument"`
}

type Step struct {
	Action     string `json:"action" jsonschema:"description=The action to be taken"`
	Motivation string `json:"motivation" jsonschema:"description=The motivation for the action"`
}
