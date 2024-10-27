// ai/reasoning/validator.go
package reasoning

import (
	"fmt"
	"reflect"
)

type Validator struct {
	chain         *ReasoningChain // Used to track validation history
	knowledgeBase *KnowledgeBase
}

func NewValidator(kb *KnowledgeBase) *Validator {
	return &Validator{
		knowledgeBase: kb,
	}
}

func (v *Validator) ValidateChain(chain *ReasoningChain) error {
	for i, step := range chain.Steps {
		if err := v.validateStep(step); err != nil {
			chain.Contradictions = append(chain.Contradictions,
				fmt.Sprintf("Step %d: %s", i, err.Error()))
		}
	}

	chain.Validated = len(chain.Contradictions) == 0
	return nil
}

func (v *Validator) validateStep(step ReasoningStep) error {
	// Validate premise
	if err := v.knowledgeBase.ValidateExpression(step.Premise); err != nil {
		return fmt.Errorf("invalid premise: %w", err)
	}

	// Validate conclusion
	if err := v.knowledgeBase.ValidateExpression(step.Conclusion); err != nil {
		return fmt.Errorf("invalid conclusion: %w", err)
	}

	// Validate logical connection between premise and conclusion
	if err := v.validateLogicalConnection(step.Premise, step.Conclusion); err != nil {
		return fmt.Errorf("invalid logical connection: %w", err)
	}

	// Validate evidence
	if err := v.validateEvidence(step.Evidence); err != nil {
		return fmt.Errorf("invalid evidence: %w", err)
	}

	return nil
}

func (v *Validator) validateLogicalConnection(premise, conclusion LogicalExpression) error {
	// Check if premise has sufficient confidence
	if premise.Confidence < 0.5 {
		return fmt.Errorf("premise confidence too low: %.2f", premise.Confidence)
	}

	// Check if operands are compatible
	for _, premiseOp := range premise.Operands {
		found := false
		for _, conclusionOp := range conclusion.Operands {
			if reflect.DeepEqual(premiseOp, conclusionOp) {
				found = true
				break
			}
		}
		if !found {
			// Track this validation in the chain
			if v.chain != nil {
				v.chain.Contradictions = append(v.chain.Contradictions,
					fmt.Sprintf("conclusion does not follow from premise: %v", premiseOp))
			}
			return fmt.Errorf("conclusion does not follow from premise: %v", premiseOp)
		}
	}

	return nil
}

func (v *Validator) validateEvidence(evidence []string) error {
	if len(evidence) == 0 {
		return fmt.Errorf("no evidence provided")
	}

	// Check each piece of evidence against the knowledge base
	for _, e := range evidence {
		if !v.knowledgeBase.HasFact(e) {
			// Track this validation in the chain
			if v.chain != nil {
				v.chain.Contradictions = append(v.chain.Contradictions,
					fmt.Sprintf("evidence not found in knowledge base: %s", e))
			}
			return fmt.Errorf("evidence not found in knowledge base: %s", e)
		}

		// Verify evidence supports the current reasoning chain
		if v.chain != nil && len(v.chain.Steps) > 0 {
			lastStep := v.chain.Steps[len(v.chain.Steps)-1]
			if !v.knowledgeBase.SupportsConclusion(e, lastStep.Conclusion) {
				v.chain.Contradictions = append(v.chain.Contradictions,
					fmt.Sprintf("evidence does not support conclusion: %s", e))
				return fmt.Errorf("evidence does not support conclusion: %s", e)
			}
		}
	}

	return nil
}

// Add helper method to KnowledgeBase
func (kb *KnowledgeBase) HasFact(fact string) bool {
	// Implementation depends on how facts are stored in the knowledge base
	return true // Placeholder
}

// Add helper method to KnowledgeBase
func (kb *KnowledgeBase) SupportsConclusion(evidence string, conclusion LogicalExpression) bool {
	// Implementation depends on how we want to verify support
	return true // Placeholder
}
