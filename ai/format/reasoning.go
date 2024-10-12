package format

type Strategy struct {
	Reasoning []Reasoning `json:"reasoning" jsonschema_description:"Dynamic composition of reasoning types you want to use" jsonschema:"anyof_ref=#/$defs/chain_of_thought;#/$defs/tree_of_thought;#/$defs/self_reflection;#/$defs/verification;#/$defs/sprintplan"`
}

type Thought string

type Reflection string

type Challenge string

type Review string

type Step struct {
	Thought    string `json:"thought" jsonschema_description:"Thought or idea at the current step in the chain"`
	Effect     string `json:"effect" jsonschema_description:"Effect of the current thought on the previous conclusions"`
	Conclusion string `json:"conclusion" jsonschema_description:"Conclusion after the current step in the chain, combined with the previous steps"`
}

type Task struct {
	Title   string `json:"title" jsonschema_description:"Title of the task"`
	Summary string `json:"summary" jsonschema_description:"Description of the task"`
	Actions []Step `json:"actions" jsonschema_description:"Sequence of actions in the task"`
}

type Story struct {
	Title   string `json:"title" jsonschema_description:"Title of the story"`
	Summary string `json:"summary" jsonschema_description:"Gherkin summary of the story"`
	Tasks   []Task `json:"tasks" jsonschema_description:"Sequence of tasks to be completed"`
}

type Epic struct {
	Title   string  `json:"title" jsonschema_description:"Title of the epic"`
	Summary string  `json:"summary" jsonschema_description:"Gherkin summary of the epic"`
	Stories []Story `json:"stories" jsonschema_description:"Sequence of stories to be completed"`
}

type Sprint struct {
	Goal    string `json:"goal" jsonschema_description:"Goal of the sprint"`
	Summary string `json:"summary" jsonschema_description:"Description of the sprint"`
	Epics   []Epic `json:"epics" jsonschema_description:"Sequence of epics to be completed"`
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
