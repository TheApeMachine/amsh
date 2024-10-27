package reasoning

// LogicalOperation represents different types of logical operations
type LogicalOperation string

const (
	AND LogicalOperation = "AND"
	OR  LogicalOperation = "OR"
	NOT LogicalOperation = "NOT"
	IF  LogicalOperation = "IF"
)

// LogicalExpression represents a formal logic structure
type LogicalExpression struct {
	Operation  LogicalOperation
	Operands   []interface{} // Can be string (atomic) or LogicalExpression
	Confidence float64
}

// MetaStrategy represents different reasoning strategies
type MetaStrategy struct {
	Name        string
	Priority    int
	Constraints []string
	Resources   map[string]float64
	Keywords    []string // Add this field for problem-specific matching
}

type ReasoningChain struct {
	Steps          []ReasoningStep
	Confidence     float64
	Validated      bool
	Contradictions []string
}

func NewReasoningChain() *ReasoningChain {
	return &ReasoningChain{
		Steps:          make([]ReasoningStep, 0),
		Contradictions: make([]string, 0),
	}
}

// Add VerificationStep type
type VerificationStep struct {
	Assumption string
	Method     string
	Result     string
	Confidence float64
}

// Update ReasoningStep to include verification fields
type ReasoningStep struct {
	Strategy            *MetaStrategy
	Premise             LogicalExpression
	Conclusion          LogicalExpression
	Confidence          float64
	Evidence            []string
	Dependencies        []string
	VerificationPrompts []string           // Add this field
	Verifications       []VerificationStep // Add this field
}
