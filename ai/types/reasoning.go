package types

// MetaStrategy represents different reasoning strategies
type MetaStrategy struct {
	Name        string
	Priority    int
	Constraints []string
	Resources   map[string]float64
	Keywords    []string
}

// ReasoningChain represents a sequence of reasoning steps
type ReasoningChain struct {
	Steps          []ReasoningStep
	Confidence     float64
	Validated      bool
	Contradictions []string
}

// ReasoningStep represents a single step in the reasoning process
type ReasoningStep struct {
	Strategy            *MetaStrategy
	Premise             LogicalExpression
	Conclusion          LogicalExpression
	Confidence          float64
	Evidence            []string
	Dependencies        []string
	VerificationPrompts []string
	Verifications       []VerificationStep
}

type LogicalOperation string

const (
	AND LogicalOperation = "AND"
	OR  LogicalOperation = "OR"
	NOT LogicalOperation = "NOT"
	IF  LogicalOperation = "IF"
)

type VerificationStep struct {
	Assumption string
	Method     string
	Result     string
	Confidence float64
}
