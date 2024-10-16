package format

import (
	"fmt"
	"strings"

	"github.com/theapemachine/amsh/utils"
)

type Strategy struct {
	Reasoning []Reasoning `json:"reasoning" jsonschema_description:"Dynamic composition of reasoning types you want to use" jsonschema:"anyof_ref=#/$defs/chain_of_thought;#/$defs/tree_of_thought;#/$defs/self_reflection;#/$defs/verification;#/$defs/sprintplan"`
}

func (s Strategy) Format() ResponseFormat {
	return s
}

func (s Strategy) String() string {
	var sb strings.Builder
	for _, reasoning := range s.Reasoning {
		sb.WriteString(reasoning.String())
		sb.WriteString("\n")
	}
	return sb.String()
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

func (s Step) String() string {
	out := utils.Muted("[STEP]\n")
	out += fmt.Sprintf(utils.Red("Thought   : %s\n"), utils.Highlight(s.Thought))
	out += fmt.Sprintf(utils.Yellow("Effect    : %s\n"), utils.Highlight(s.Effect))
	out += fmt.Sprintf(utils.Green("Conclusion: %s\n"), utils.Highlight(s.Conclusion))
	out += utils.Muted("[/STEP]\n")
	return out
}

type Task struct {
	Title   string `json:"title" jsonschema_description:"Title of the task"`
	Summary string `json:"summary" jsonschema_description:"Description of the task"`
	Actions []Step `json:"actions" jsonschema_description:"Sequence of actions in the task"`
}

func (t Task) String() string {
	out := utils.Muted("[TASK]\n")
	out += fmt.Sprintf(utils.Red("Title     : %s\n"), utils.Highlight(t.Title))
	out += fmt.Sprintf(utils.Yellow("Summary   : %s\n"), utils.Highlight(t.Summary))
	for _, action := range t.Actions {
		out += action.String()
		out += "\n"
	}
	out += utils.Muted("[/TASK]\n")
	return out
}

type Story struct {
	Title   string `json:"title" jsonschema_description:"Title of the story"`
	Summary string `json:"summary" jsonschema_description:"Gherkin summary of the story"`
	Tasks   []Task `json:"tasks" jsonschema_description:"Sequence of tasks to be completed"`
}

func (s Story) String() string {
	out := utils.Muted("[STORY]\n")
	out += fmt.Sprintf(utils.Red("Title     : %s\n"), utils.Highlight(s.Title))
	out += fmt.Sprintf(utils.Yellow("Summary   : %s\n"), utils.Highlight(s.Summary))
	for _, task := range s.Tasks {
		out += task.String()
		out += "\n"
	}
	out += utils.Muted("[/STORY]\n")
	return out
}

type Epic struct {
	Title   string  `json:"title" jsonschema_description:"Title of the epic"`
	Summary string  `json:"summary" jsonschema_description:"Gherkin summary of the epic"`
	Stories []Story `json:"stories" jsonschema_description:"Sequence of stories to be completed"`
}

func (e Epic) String() string {
	out := utils.Muted("[EPIC]\n")
	out += fmt.Sprintf(utils.Red("Title     : %s\n"), utils.Highlight(e.Title))
	out += fmt.Sprintf(utils.Yellow("Summary   : %s\n"), utils.Highlight(e.Summary))
	for _, story := range e.Stories {
		out += story.String()
		out += "\n"
	}
	out += utils.Muted("[/EPIC]\n")
	return out
}

type Sprint struct {
	Goal    string `json:"goal" jsonschema_description:"Goal of the sprint"`
	Summary string `json:"summary" jsonschema_description:"Description of the sprint"`
	Epics   []Epic `json:"epics" jsonschema_description:"Sequence of epics to be completed"`
}

func (s Sprint) String() string {
	out := utils.Muted("[SPRINT]\n")
	out += fmt.Sprintf(utils.Red("Goal      : %s\n"), utils.Highlight(s.Goal))
	out += fmt.Sprintf(utils.Yellow("Summary   : %s\n"), utils.Highlight(s.Summary))
	for _, epic := range s.Epics {
		out += epic.String()
		out += "\n"
	}
	out += utils.Muted("[/SPRINT]\n")
	return out
}

type Reasoning struct {
	ChainOfThought ChainOfThought `json:"chain_of_thought" jsonschema_description:"A chain of thoughts"`
	TreeOfThought  TreeOfThought  `json:"tree_of_thought" jsonschema_description:"A tree of thoughts"`
	SelfReflection SelfReflection `json:"self_reflection" jsonschema_description:"A self reflection"`
	Verification   Verification   `json:"verification" jsonschema_description:"A verification"`
}

func (r Reasoning) String() string {
	out := utils.Muted("[REASONING]\n")
	out += fmt.Sprintf(utils.Red("Chain of Thought: %s\n"), utils.Highlight(r.ChainOfThought.String()))
	if len(r.ChainOfThought.Steps) > 0 {
		out += r.ChainOfThought.String()
		out += "\n"
	}
	if r.TreeOfThought.RootThought != "" {
		out += r.TreeOfThought.String()
		out += "\n"
	}
	if len(r.SelfReflection.Reflections) > 0 {
		out += r.SelfReflection.String()
		out += "\n"
	}
	if r.Verification.Target != "" {
		out += r.Verification.String()
		out += "\n"
	}
	out += utils.Muted("[/REASONING]\n")
	return out
}

type ChainOfThought struct {
	Steps       []Step `json:"steps" jsonschema_description:"Sequence of steps in the chain of thought"`
	FinalAnswer string `json:"final_answer" jsonschema_description:"The final conclusion or answer"`
}

func (c ChainOfThought) String() string {
	out := utils.Muted("[CHAIN OF THOUGH]\n")
	out += fmt.Sprintf(utils.Red("Final Answer: %s\n"), utils.Highlight(c.FinalAnswer))
	for _, step := range c.Steps {
		out += step.String()
		out += "\n"
	}
	out += fmt.Sprintf("Final Answer: %s", c.FinalAnswer)
	return out
}

type TreeOfThought struct {
	RootThought Thought    `json:"root_thought" jsonschema_description:"The root thought of the tree"`
	Branches    []*Thought `json:"branches" jsonschema_description:"Branches in the tree of thought" jsonschema:"anyof_ref=#/$defs/thought"`
	FinalAnswer string     `json:"final_answer" jsonschema_description:"The final conclusion or answer"`
}

func (t TreeOfThought) String() string {
	out := utils.Muted("[TREE OF THOUGH]\n")
	out += fmt.Sprintf(utils.Red("Root Thought: %s\n"), utils.Highlight(string(t.RootThought)))
	for _, branch := range t.Branches {
		if branch != nil {
			out += fmt.Sprintf("Branch: %s\n", *branch)
		}
	}
	out += fmt.Sprintf("Final Answer: %s", t.FinalAnswer)
	return out
}

type SelfReflection struct {
	Reflections []Reflection `json:"reflections" jsonschema_description:"Series of reflections"`
	FinalAnswer string       `json:"final_answer" jsonschema_description:"The final conclusion or answer"`
}

func (s SelfReflection) String() string {
	out := utils.Muted("[SELF REFLECTION]\n")
	out += fmt.Sprintf(utils.Red("Final Answer: %s\n"), utils.Highlight(s.FinalAnswer))
	for _, reflection := range s.Reflections {
		out += fmt.Sprintf("Reflection: %s\n", reflection)
	}
	out += fmt.Sprintf("Final Answer: %s", s.FinalAnswer)
	return out
}

type Verification struct {
	Target     string      `json:"target" jsonschema_description:"A previous conclusion, hypothesis, prediction, etc. to be verified"`
	Challenges []Challenge `json:"challenges" jsonschema_description:"Challenges posed against the current belief or conclusion"`
	Review     []Review    `json:"review" jsonschema_description:"Review of the challenges to reach a final decision"`
	Result     string      `json:"result" jsonschema_description:"Result of the verification process"`
}

func (v Verification) String() string {
	out := utils.Muted("[VERIFICATION]\n")
	out += fmt.Sprintf(utils.Red("Target: %s\n"), utils.Highlight(v.Target))
	for _, challenge := range v.Challenges {
		out += fmt.Sprintf("Challenge: %s\n", challenge)
	}
	for _, review := range v.Review {
		out += fmt.Sprintf("Review: %s\n", review)
	}
	out += fmt.Sprintf("Result: %s", v.Result)
	return out
}
