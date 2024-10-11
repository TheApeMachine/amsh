package format

type Strategy struct {
	Reasoning []Reasoning `json:"reasoning" jsonschema_description:"Dynamic composition of reasoning types you want to use" jsonschema:"anyof_ref=#/$defs/thought;#/$defs/reflection;#/$defs/challenge;#/$defs/review;#/$defs/chain_of_thought;#/$defs/tree_of_thought;#/$defs/self_reflection;#/$defs/verification"`
}

type Thought string

type Reflection string

type Challenge string

type Review string

type Step struct {
	Description string `json:"description" jsonschema_description:"Description of the step"`
	Action      string `json:"action" jsonschema_description:"Action taken in this step"`
	Result      string `json:"result" jsonschema_description:"Result of the action"`
}

type Reasoning struct {
	ChainOfThought ChainOfThought `json:"chain_of_thought" jsonschema_description:"A chain of thoughts"`
	TreeOfThought  TreeOfThought  `json:"tree_of_thought" jsonschema_description:"A tree of thoughts"`
	SelfReflection SelfReflection `json:"self_reflection" jsonschema_description:"A self reflection"`
	Verification   Verification   `json:"verification" jsonschema_description:"A verification"`
}

type ChainOfThought struct {
	Steps       []Step `json:"steps" jsonschema_description:"Sequence of steps in the chain of thought"`
	FinalAnswer string `json:"final_answer" jsonschema_description:"The final conclusion or answer"`
}

type TreeOfThought struct {
	RootThought Thought    `json:"root_thought" jsonschema_description:"The root thought of the tree"`
	Branches    []*Thought `json:"branches" jsonschema_description:"Branches in the tree of thought" jsonschema:"anyof_ref=#/$defs/thought"`
	FinalAnswer string     `json:"final_answer" jsonschema_description:"The final conclusion or answer"`
}

type SelfReflection struct {
	Reflections []Reflection `json:"reflections" jsonschema_description:"Series of reflections"`
	FinalAnswer string       `json:"final_answer" jsonschema_description:"The final conclusion or answer"`
}

type Verification struct {
	Target     string      `json:"target" jsonschema_description:"A previous conclusion, hypothesis, prediction, etc. to be verified"`
	Challenges []Challenge `json:"challenges" jsonschema_description:"Challenges posed against the current belief or conclusion"`
	Review     []Review    `json:"review" jsonschema_description:"Review of the challenges to reach a final decision"`
	Result     string      `json:"result" jsonschema_description:"Result of the verification process"`
}
