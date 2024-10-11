package format

import (
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

var StrategyMap = map[string]*Response{
	"reasoning_strategy":      NewResponse("reasoning_strategy", NewReasoningStrategy()),
	"self_reflection":         NewResponse("self_reflection", NewSelfReflection()),
	"chain_of_thought":        NewResponse("chain_of_thought", NewChainOfThought()),
	"tree_of_thought":         NewResponse("tree_of_thought", NewTreeOfThought()),
	"first_principles":        NewResponse("first_principles", NewFirstPrinciplesReasoning()),
	"step_by_step_execution":  NewResponse("step_by_step_execution", NewStepByStepExecution()),
	"pros_and_cons_analysis":  NewResponse("pros_and_cons_analysis", NewProsAndConsAnalysis()),
	"divide_and_conquer":      NewResponse("divide_and_conquer", NewDivideAndConquer()),
	"hypothesis_testing":      NewResponse("hypothesis_testing", NewHypothesisTesting()),
	"risk_assessment":         NewResponse("risk_assessment", NewRiskAssessment()),
	"backwards_reasoning":     NewResponse("backwards_reasoning", NewBackwardReasoning()),
	"roleplay_simulation":     NewResponse("roleplay_simulation", NewRolePlaySimulation()),
	"scenario_analysis":       NewResponse("scenario_analysis", NewScenarioAnalysis()),
	"counterfactual_thinking": NewResponse("counterfactual", NewCounterfactualThinking()),
	"final":                   NewResponse("final", NewFinal()),
	"environment_interaction": NewResponse("environment_interaction", NewEnvironmentInteraction()),
}

type ReasoningStrategy struct {
	Steps []struct {
		Thought       string `json:"thought"`
		Strategy      string `json:"strategy"`
		Clarification string `json:"clarification"`
	} `json:"steps"`
	OrderedStrategies []string `json:"ordered_strategies"`
}

func NewReasoningStrategy() *ReasoningStrategy {
	errnie.Trace()
	return &ReasoningStrategy{}
}

func (strategy *ReasoningStrategy) FinalAnswer() string {
	return strings.Join(strategy.OrderedStrategies, ",")
}

func (strategy *ReasoningStrategy) Schema() (*jsonschema.Definition, error) {
	return jsonschema.GenerateSchemaForType(strategy)
}

func (strategy *ReasoningStrategy) ToString() string {
	out := []string{}
	out = append(out, dark("  [REASONING STRATEGY]"))
	for _, step := range strategy.Steps {
		out = append(out, muted("    [STEP]"))
		out = append(out, red("            Thought: ")+highlight(step.Thought))
		out = append(out, yellow("           Strategy: ")+highlight(step.Strategy))
		out = append(out, green("      Clarification: ")+highlight(step.Clarification))
		out = append(out, muted("    [/STEP]"))
	}
	out = append(out, blue("    OrderedStrategies: ")+highlight(strings.Join(strategy.OrderedStrategies, ", ")))
	out = append(out, dark("  [/REASONING STRATEGY]"))
	return strings.Join(out, "\n")
}
