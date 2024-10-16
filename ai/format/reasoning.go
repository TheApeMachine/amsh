package format

import (
	"fmt"
	"strings"

	"github.com/theapemachine/amsh/utils"
)

type Reasoning struct {
	Reasoning []ReasoningType `json:"reasoning"`
	Done      bool            `json:"done"`
}

func (r Reasoning) String() string {
	var sb strings.Builder
	sb.WriteString(utils.Muted("[REASONING]\n"))
	for _, reasoning := range r.Reasoning {
		sb.WriteString(reasoning.String())
		sb.WriteString("\n")
	}
	sb.WriteString(utils.Muted("[/REASONING]\n"))
	return sb.String()
}

type ReasoningType struct {
	ChainOfThought         *ChainOfThought         `json:"chain_of_thought,omitempty"`
	TreeOfThought          *TreeOfThought          `json:"tree_of_thought,omitempty"`
	SelfReflection         *SelfReflection         `json:"self_reflection,omitempty"`
	Verification           *Verification           `json:"verification,omitempty"`
	CustomReasoning        *CustomReasoning        `json:"custom_reasoning,omitempty"`
	SprintPlan             *SprintPlan             `json:"sprint_plan,omitempty"`
	Counterfactual         *Counterfactual         `json:"counterfactual,omitempty"`
	MetaReasoning          *MetaReasoning          `json:"meta_reasoning,omitempty"`
	CollaborativeReasoning *CollaborativeReasoning `json:"collaborative_reasoning,omitempty"`
	Roleplay               *Roleplay               `json:"roleplay,omitempty"`
	Debate                 *Debate                 `json:"debate,omitempty"`
	Exploration            *Exploration            `json:"exploration,omitempty"`
}

func (r ReasoningType) String() string {
	out := utils.Muted("[REASONING]\n")
	if r.ChainOfThought != nil {
		out += r.ChainOfThought.String() + "\n"
	}
	if r.TreeOfThought != nil {
		out += r.TreeOfThought.String() + "\n"
	}
	if r.SelfReflection != nil {
		out += r.SelfReflection.String() + "\n"
	}
	if r.Verification != nil {
		out += r.Verification.String() + "\n"
	}
	if r.CustomReasoning != nil {
		out += r.CustomReasoning.String() + "\n"
	}
	if r.SprintPlan != nil {
		out += r.SprintPlan.String() + "\n"
	}
	if r.Counterfactual != nil {
		out += r.Counterfactual.String() + "\n"
	}
	if r.MetaReasoning != nil {
		out += r.MetaReasoning.String() + "\n"
	}
	if r.CollaborativeReasoning != nil {
		out += r.CollaborativeReasoning.String() + "\n"
	}
	if r.Roleplay != nil {
		out += r.Roleplay.String() + "\n"
	}
	if r.Debate != nil {
		out += r.Debate.String() + "\n"
	}
	if r.Exploration != nil {
		out += r.Exploration.String() + "\n"
	}
	out += utils.Muted("[/REASONING]\n")
	return out
}

// Chain of Thought reasoning type
type ChainOfThought struct {
	Steps       []Step `json:"steps"`
	FinalAnswer string `json:"final_answer"`
}

func (c ChainOfThought) String() string {
	out := utils.Muted("[CHAIN OF THOUGHT]\n")
	for _, step := range c.Steps {
		out += step.String() + "\n"
	}
	out += fmt.Sprintf(utils.Red("Final Answer: %s\n"), utils.Highlight(c.FinalAnswer))
	out += utils.Muted("[/CHAIN OF THOUGHT]\n")
	return out
}

// Tree of Thought reasoning type
type TreeOfThought struct {
	RootThought Thought    `json:"root_thought"`
	Branches    []*Thought `json:"branches,omitempty"`
	FinalAnswer string     `json:"final_answer"`
}

func (t TreeOfThought) String() string {
	out := utils.Muted("[TREE OF THOUGHT]\n")
	out += fmt.Sprintf(utils.Red("Root Thought: %s\n"), utils.Highlight(string(t.RootThought)))
	for _, branch := range t.Branches {
		if branch != nil {
			out += fmt.Sprintf(utils.Yellow("Branch: %s\n"), *branch)
		}
	}
	out += fmt.Sprintf(utils.Green("Final Answer: %s\n"), utils.Highlight(t.FinalAnswer))
	out += utils.Muted("[/TREE OF THOUGHT]\n")
	return out
}

// Self Reflection reasoning type
type SelfReflection struct {
	Reflections []Reflection `json:"reflections"`
	FinalAnswer string       `json:"final_answer"`
}

func (s SelfReflection) String() string {
	out := utils.Muted("[SELF REFLECTION]\n")
	for _, reflection := range s.Reflections {
		out += fmt.Sprintf(utils.Yellow("Reflection: %s\n"), reflection)
	}
	out += fmt.Sprintf(utils.Green("Final Answer: %s\n"), utils.Highlight(s.FinalAnswer))
	out += utils.Muted("[/SELF REFLECTION]\n")
	return out
}

// Verification reasoning type
type Verification struct {
	Target     string      `json:"target"`
	Challenges []Challenge `json:"challenges"`
	Review     []Review    `json:"review"`
	Result     string      `json:"result"`
}

func (v Verification) String() string {
	out := utils.Muted("[VERIFICATION]\n")
	out += fmt.Sprintf(utils.Red("Target: %s\n"), utils.Highlight(v.Target))
	for _, challenge := range v.Challenges {
		out += fmt.Sprintf(utils.Yellow("Challenge: %s\n"), challenge)
	}
	for _, review := range v.Review {
		out += fmt.Sprintf(utils.Green("Review: %s\n"), review)
	}
	out += fmt.Sprintf(utils.Green("Result: %s\n"), utils.Highlight(v.Result))
	out += utils.Muted("[/VERIFICATION]\n")
	return out
}

// Custom Reasoning
type CustomReasoning struct {
	Description string `json:"description"`
	Steps       []Step `json:"steps"`
}

func (c CustomReasoning) String() string {
	out := utils.Muted("[CUSTOM REASONING]\n")
	out += fmt.Sprintf(utils.Red("Description: %s\n"), utils.Highlight(c.Description))
	for _, step := range c.Steps {
		out += step.String() + "\n"
	}
	out += utils.Muted("[/CUSTOM REASONING]\n")
	return out
}

// Sprint Plan reasoning type
type SprintPlan struct {
	Goal        string `json:"goal"`
	Tasks       []Task `json:"tasks"`
	FinalStatus string `json:"final_status"`
}

func (s SprintPlan) String() string {
	out := utils.Muted("[SPRINT PLAN]\n")
	out += fmt.Sprintf(utils.Red("Goal: %s\n"), utils.Highlight(s.Goal))
	for _, task := range s.Tasks {
		out += task.String() + "\n"
	}
	out += fmt.Sprintf(utils.Green("Final Status: %s\n"), utils.Highlight(s.FinalStatus))
	out += utils.Muted("[/SPRINT PLAN]\n")
	return out
}

type Task struct {
	Title string `json:"title"`
	Steps []Step `json:"steps"`
}

func (t Task) String() string {
	out := utils.Muted("[TASK]\n")
	out += fmt.Sprintf(utils.Red("Title: %s\n"), utils.Highlight(t.Title))
	for _, step := range t.Steps {
		out += step.String() + "\n"
	}
	out += utils.Muted("[/TASK]\n")
	return out
}

// Counterfactual reasoning type
type Counterfactual struct {
	Hypothesis   string `json:"hypothesis"`
	AlteredSteps []Step `json:"altered_steps"`
	Outcome      string `json:"outcome"`
}

func (c Counterfactual) String() string {
	out := utils.Muted("[COUNTERFACTUAL]\n")
	out += fmt.Sprintf(utils.Red("Hypothesis: %s\n"), utils.Highlight(c.Hypothesis))
	for _, step := range c.AlteredSteps {
		out += step.String() + "\n"
	}
	out += fmt.Sprintf(utils.Green("Outcome: %s\n"), utils.Highlight(c.Outcome))
	out += utils.Muted("[/COUNTERFACTUAL]\n")
	return out
}

// Meta Reasoning type
type MetaReasoning struct {
	Strategy     string       `json:"strategy"`
	Observations []string     `json:"observations"`
	Reflections  []Reflection `json:"reflections"`
}

func (m MetaReasoning) String() string {
	out := utils.Muted("[META REASONING]\n")
	out += fmt.Sprintf(utils.Red("Strategy: %s\n"), utils.Highlight(m.Strategy))
	for _, observation := range m.Observations {
		out += fmt.Sprintf(utils.Yellow("Observation: %s\n"), observation)
	}
	for _, reflection := range m.Reflections {
		out += fmt.Sprintf(utils.Green("Reflection: %s\n"), reflection)
	}
	out += utils.Muted("[/META REASONING]\n")
	return out
}

// Collaborative Reasoning type
type CollaborativeReasoning struct {
	Participants []string `json:"participants"`
	Steps        []Step   `json:"steps"`
	Consensus    string   `json:"consensus"`
}

func (c CollaborativeReasoning) String() string {
	out := utils.Muted("[COLLABORATIVE REASONING]\n")
	out += fmt.Sprintf(utils.Red("Participants: %s\n"), utils.Highlight(strings.Join(c.Participants, ", ")))
	for _, step := range c.Steps {
		out += step.String() + "\n"
	}
	out += fmt.Sprintf(utils.Green("Consensus: %s\n"), utils.Highlight(c.Consensus))
	out += utils.Muted("[/COLLABORATIVE REASONING]\n")
	return out
}

// Roleplay reasoning type
type Roleplay struct {
	Roles    []string `json:"roles"`
	Scenario string   `json:"scenario"`
	Outcome  string   `json:"outcome"`
}

func (r Roleplay) String() string {
	out := utils.Muted("[ROLEPLAY]\n")
	out += fmt.Sprintf(utils.Red("Roles: %s\n"), utils.Highlight(strings.Join(r.Roles, ", ")))
	out += fmt.Sprintf(utils.Yellow("Scenario: %s\n"), utils.Highlight(r.Scenario))
	out += fmt.Sprintf(utils.Green("Outcome: %s\n"), utils.Highlight(r.Outcome))
	out += utils.Muted("[/ROLEPLAY]\n")
	return out
}

// Debate reasoning type
type Debate struct {
	Topic        string   `json:"topic"`
	ProArguments []string `json:"pro_arguments"`
	ConArguments []string `json:"con_arguments"`
	Conclusion   string   `json:"conclusion"`
}

func (d Debate) String() string {
	out := utils.Muted("[DEBATE]\n")
	out += fmt.Sprintf(utils.Red("Topic: %s\n"), utils.Highlight(d.Topic))
	out += fmt.Sprintf(utils.Yellow("Pro Arguments: %s\n"), utils.Highlight(strings.Join(d.ProArguments, ", ")))
	out += fmt.Sprintf(utils.Yellow("Con Arguments: %s\n"), utils.Highlight(strings.Join(d.ConArguments, ", ")))
	out += fmt.Sprintf(utils.Green("Conclusion: %s\n"), utils.Highlight(d.Conclusion))
	out += utils.Muted("[/DEBATE]\n")
	return out
}

// Exploration reasoning type
type Exploration struct {
	StartingPoint string   `json:"starting_point"`
	Discoveries   []string `json:"discoveries"`
	Conclusion    string   `json:"conclusion"`
}

func (e Exploration) String() string {
	out := utils.Muted("[EXPLORATION]\n")
	out += fmt.Sprintf(utils.Red("Starting Point: %s\n"), utils.Highlight(e.StartingPoint))
	for _, discovery := range e.Discoveries {
		out += fmt.Sprintf(utils.Yellow("Discovery: %s\n"), discovery)
	}
	out += fmt.Sprintf(utils.Green("Conclusion: %s\n"), utils.Highlight(e.Conclusion))
	out += utils.Muted("[/EXPLORATION]\n")
	return out
}

// Definitions of the string types used in reasoning
type Thought string
type Reflection string
type Challenge string
type Review string

// Step structure used by several reasoning types
type Step struct {
	Thought    string `json:"thought"`
	Effect     string `json:"effect"`
	Conclusion string `json:"conclusion"`
}

func (s Step) String() string {
	out := utils.Muted("[STEP]\n")
	out += fmt.Sprintf(utils.Red("Thought   : %s\n"), utils.Highlight(s.Thought))
	out += fmt.Sprintf(utils.Yellow("Effect    : %s\n"), utils.Highlight(s.Effect))
	out += fmt.Sprintf(utils.Green("Conclusion: %s\n"), utils.Highlight(s.Conclusion))
	out += utils.Muted("[/STEP]\n")
	return out
}
