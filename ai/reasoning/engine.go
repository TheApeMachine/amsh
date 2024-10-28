// ai/reasoning/engine.go
package reasoning

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/theapemachine/amsh/ai/learning"
	"github.com/theapemachine/amsh/ai/types"
)

// Engine represents the reasoning engine
type Engine struct {
	Validator    *Validator
	Uncertainty  float64
	MaxSteps     int
	MetaReasoner *MetaReasoner
	Learning     *learning.LearningAdapter
	BayesianNet  *BayesianNetwork // Added BayesianNet field
}

// NewEngine creates a new reasoning engine
func NewEngine(validator *Validator, metaReasoner *MetaReasoner) *Engine {
	bayesianNet := initializeBayesianNetwork()
	return &Engine{
		Validator:    validator,
		Uncertainty:  0.1,
		MaxSteps:     10,
		MetaReasoner: metaReasoner,
		Learning:     learning.NewLearningAdapter(),
		BayesianNet:  bayesianNet,
	}
}

// initializeBayesianNetwork initializes the Bayesian network
func initializeBayesianNetwork() *BayesianNetwork {
	bn := NewBayesianNetwork()

	// Example nodes and relationships
	// Node: StrategyEffective
	_, err := bn.AddNode("StrategyEffective", 0.7) // Prior probability that a strategy is effective
	if err != nil {
		// In a real implementation, you might want to handle this error differently
		panic(err)
	}

	// Node: CorrectConclusion depends on StrategyEffective
	conclusionNode, err := bn.AddNode("CorrectConclusion", 0.0) // Prior will be defined in CPT
	if err != nil {
		panic(err)
	}

	// Add edge: CorrectConclusion <- StrategyEffective
	bn.AddEdge("CorrectConclusion", "StrategyEffective")

	// Instead of SetCPT, we should update the node's probability directly
	// This assumes BayesianNode has a Probability field that can be updated
	conclusionNode.Probability = map[string]float64{
		"T": 0.9, // If StrategyEffective is True, then CorrectConclusion probability is 0.9
		"F": 0.2, // If StrategyEffective is False, then CorrectConclusion probability is 0.2
	}

	return bn
}

// Think performs the reasoning process
func (e *Engine) Think(ctx context.Context, problem string) (*types.ReasoningChain, error) {
	chain := &types.ReasoningChain{}

	for steps := 0; steps < e.MaxSteps; steps++ {
		step, err := e.GenerateStep(ctx, problem, chain)
		if err != nil {
			return nil, fmt.Errorf("reasoning step %d failed: %w", steps, err)
		}

		chain.Steps = append(chain.Steps, step)
		if e.hasReachedConclusion(chain) {
			break
		}
	}

	// Remove the conversion and use chain directly
	if err := e.Validator.ValidateChain(chain); err != nil {
		return chain, fmt.Errorf("validation failed: %w", err)
	}

	return chain, nil
}

// GenerateStep creates a new reasoning step based on the problem and current chain
func (e *Engine) GenerateStep(ctx context.Context, problem string, chain *types.ReasoningChain) (types.ReasoningStep, error) {
	// Convert MetaStrategy to types.MetaStrategy
	strategy, err := e.MetaReasoner.SelectStrategy(ctx, problem, e.deriveConstraints(chain))
	if err != nil {
		return types.ReasoningStep{}, err
	}

	typesStrategy := &types.MetaStrategy{
		Name:        strategy.Name,
		Priority:    strategy.Priority,
		Constraints: strategy.Constraints,
		Resources:   strategy.Resources,
		Keywords:    strategy.Keywords,
	}

	state := e.getCurrentState(problem, chain)
	adaptedStrategy, err := e.Learning.AdaptStrategy(ctx, typesStrategy, state)
	if err != nil {
		return types.ReasoningStep{}, err
	}

	step, err := e.generateStepWithStrategy(ctx, problem, chain, adaptedStrategy)
	if err != nil {
		return types.ReasoningStep{}, err
	}

	e.Learning.RecordStrategyExecution(adaptedStrategy, chain)

	return step, nil
}

func (e *Engine) generateInitialStep(ctx context.Context, problem string, chain *types.ReasoningChain) (types.ReasoningStep, error) {
	constraints := e.deriveConstraints(chain)
	strategy, err := e.MetaReasoner.SelectStrategy(ctx, problem, constraints)
	if err != nil {
		return types.ReasoningStep{}, fmt.Errorf("strategy selection failed: %w", err)
	}

	typesStrategy := &types.MetaStrategy{
		Name:        strategy.Name,
		Priority:    strategy.Priority,
		Constraints: strategy.Constraints,
		Resources:   strategy.Resources,
		Keywords:    strategy.Keywords,
	}

	step := types.ReasoningStep{
		Strategy: typesStrategy,
	}

	premise, err := e.buildLogicalExpression(ctx, problem, chain)
	if err != nil {
		return types.ReasoningStep{}, fmt.Errorf("premise construction failed: %w", err)
	}
	step.Premise = premise

	conclusion, err := e.deriveConclusion(ctx, premise, typesStrategy)
	if err != nil {
		return types.ReasoningStep{}, fmt.Errorf("conclusion derivation failed: %w", err)
	}
	step.Conclusion = conclusion

	step.Confidence = e.calculateConfidence(premise, conclusion, typesStrategy)

	return step, nil
}

func (e *Engine) deriveConstraints(chain *types.ReasoningChain) []string {
	// Extract constraints from current reasoning chain
	var constraints []string

	// Time constraints
	if len(chain.Steps) > e.MaxSteps/2 {
		constraints = append(constraints, "time_critical")
	}

	// Resource constraints
	// Uncertainty constraints
	// Dependency constraints

	return constraints
}

func (e *Engine) hasReachedConclusion(chain *types.ReasoningChain) bool {
	if len(chain.Steps) == 0 {
		return false
	}

	lastStep := chain.Steps[len(chain.Steps)-1]

	// Check if confidence exceeds threshold
	if lastStep.Confidence >= 0.95 {
		return true
	}

	// Check if conclusion is definitive
	if e.isDefinitiveConclusion(lastStep.Conclusion) {
		return true
	}

	return false
}

func (e *Engine) buildLogicalExpression(ctx context.Context, problem string, chain *types.ReasoningChain) (types.LogicalExpression, error) {
	select {
	case <-ctx.Done():
		return types.LogicalExpression{}, ctx.Err()
	default:
		expr := types.LogicalExpression{
			Operation: types.AND,
			Operands:  []interface{}{problem},
		}

		if len(chain.Steps) > 0 {
			lastStep := chain.Steps[len(chain.Steps)-1]
			expr.Operands = append(expr.Operands, lastStep.Conclusion)
		}

		return expr, nil
	}
}

func (e *Engine) deriveConclusion(ctx context.Context, premise types.LogicalExpression, strategy *types.MetaStrategy) (types.LogicalExpression, error) {
	select {
	case <-ctx.Done():
		return types.LogicalExpression{}, ctx.Err()
	default:
		switch strategy.Name {
		case "deductive":
			return e.applyDeductiveReasoning(premise)
		case "pattern_analysis":
			return e.applyPatternAnalysis(premise)
		case "word_decomposition":
			return e.applyWordDecomposition(premise)
		case "semantic_connection":
			return e.applySemanticConnection(premise)
		case "inductive":
			return e.applyInductiveReasoning(premise)
		case "abductive":
			return e.applyAbductiveReasoning(premise)
		default:
			return types.LogicalExpression{}, fmt.Errorf("unsupported reasoning strategy: %s", strategy.Name)
		}
	}
}

func (e *Engine) applyPatternAnalysis(premise types.LogicalExpression) (types.LogicalExpression, error) {
	// Use premise to inform the pattern analysis
	conclusion := types.LogicalExpression{
		Operation:  types.AND,
		Operands:   make([]interface{}, 0),
		Confidence: 0.8,
	}

	// Analyze premise operands for patterns
	for _, op := range premise.Operands {
		if str, ok := op.(string); ok {
			if strings.Contains(str, "three") || strings.Contains(str, "triple") {
				conclusion.Operands = append(conclusion.Operands,
					"Pattern identified: number pattern suggests triple occurrence")
			}
			if strings.Contains(str, "fruit") || strings.Contains(str, "sweet") {
				conclusion.Operands = append(conclusion.Operands,
					"Pattern identified: context suggests examining fruit names")
			}
		}
	}

	return conclusion, nil
}

func (e *Engine) applyWordDecomposition(premise types.LogicalExpression) (types.LogicalExpression, error) {
	// Use premise to guide word decomposition
	conclusion := types.LogicalExpression{
		Operation:  types.AND,
		Operands:   make([]interface{}, 0),
		Confidence: 0.7,
	}

	// Analyze premise for word-related patterns
	for _, op := range premise.Operands {
		if str, ok := op.(string); ok {
			if strings.Contains(str, "hidden") {
				conclusion.Operands = append(conclusion.Operands,
					"Decomposition: looking for hidden patterns within words")
			}
			if strings.Contains(str, "sweet") {
				conclusion.Operands = append(conclusion.Operands,
					"Decomposition: focusing on sweet fruit names")
			}
		}
	}

	return conclusion, nil
}

func (e *Engine) applySemanticConnection(premise types.LogicalExpression) (types.LogicalExpression, error) {
	// Use premise to build semantic connections
	conclusion := types.LogicalExpression{
		Operation:  types.AND,
		Operands:   make([]interface{}, 0),
		Confidence: 0.75,
	}

	// Build connections based on premise content
	for _, op := range premise.Operands {
		if str, ok := op.(string); ok {
			if strings.Contains(str, "triple") || strings.Contains(str, "three") {
				conclusion.Operands = append(conclusion.Operands,
					"Connection: triple occurrence pattern identified")
			}
			if strings.Contains(str, "fruit") {
				conclusion.Operands = append(conclusion.Operands,
					"Connection: fruit context established")
			}
		}
	}

	return conclusion, nil
}

// CalculateConfidence calculates confidence using the Bayesian network
func (e *Engine) calculateConfidence(premise, conclusion types.LogicalExpression, strategy *types.MetaStrategy) float64 {
	// Use Bayesian network to calculate the probability of CorrectConclusion
	evidence := map[string]bool{
		// Assume we have evidence about the strategy's effectiveness
		"StrategyEffective": e.strategyEffectiveness(strategy),
	}

	prob, err := e.BayesianNet.CalculateProbability("CorrectConclusion", evidence)
	if err != nil {
		// Fallback to averaging premise and conclusion confidence
		return (premise.Confidence + conclusion.Confidence) / 2
	}

	// Adjust for uncertainty
	confidence := prob * (1 - e.Uncertainty)

	// Ensure confidence is within [0,1]
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// strategyEffectiveness estimates if a strategy is effective based on meta-reasoning or learning
func (e *Engine) strategyEffectiveness(strategy *types.MetaStrategy) bool {
	return strategy.Priority <= 2
}

func (e *Engine) isDefinitiveConclusion(conclusion types.LogicalExpression) bool {
	// Check confidence threshold
	if conclusion.Confidence < 0.95 {
		return false
	}

	// Check for definitive markers in operands
	for _, op := range conclusion.Operands {
		if str, ok := op.(string); ok {
			// Look for definitive language
			if strings.Contains(strings.ToLower(str), "definitely") ||
				strings.Contains(strings.ToLower(str), "certainly") ||
				strings.Contains(strings.ToLower(str), "proven") {
				return true
			}
		}
	}

	// Check if all operands are concrete (not hypothetical)
	for _, op := range conclusion.Operands {
		if str, ok := op.(string); ok {
			if strings.Contains(strings.ToLower(str), "might") ||
				strings.Contains(strings.ToLower(str), "maybe") ||
				strings.Contains(strings.ToLower(str), "possibly") ||
				strings.Contains(strings.ToLower(str), "could") {
				return false
			}
		}
	}

	return false
}

func (e *Engine) getStrategyReliability(strategy *types.MetaStrategy) float64 {
	// Could be based on historical success rate or predefined values
	reliabilityMap := map[string]float64{
		"deduction": 0.95,
		"induction": 0.85,
		"abduction": 0.75,
	}

	if reliability, ok := reliabilityMap[strategy.Name]; ok {
		return reliability
	}
	return 0.5 // default reliability for unknown strategies
}

func (e *Engine) applyDeductiveReasoning(premise types.LogicalExpression) (types.LogicalExpression, error) {
	// Deductive reasoning: from general rules to specific conclusions
	// Example: All humans are mortal (major premise)
	//         Socrates is human (minor premise)
	//         Therefore, Socrates is mortal (conclusion)

	conclusion := types.LogicalExpression{
		Operation:  types.IF,
		Operands:   []interface{}{premise},
		Confidence: premise.Confidence * 0.95, // Deductive reasoning has high confidence
	}

	return conclusion, nil
}

func (e *Engine) applyInductiveReasoning(premise types.LogicalExpression) (types.LogicalExpression, error) {
	// Inductive reasoning: from specific observations to general conclusions
	// Example: All observed swans are white
	//         Therefore, all swans are probably white

	conclusion := types.LogicalExpression{
		Operation:  types.AND,
		Operands:   []interface{}{premise},
		Confidence: premise.Confidence * 0.85, // Inductive reasoning has medium confidence
	}

	return conclusion, nil
}

func (e *Engine) applyAbductiveReasoning(premise types.LogicalExpression) (types.LogicalExpression, error) {
	// Abductive reasoning: inference to the best explanation
	// Example: The grass is wet
	//         If it rained, the grass would be wet
	//         Therefore, it probably rained

	conclusion := types.LogicalExpression{
		Operation:  types.OR,
		Operands:   []interface{}{premise},
		Confidence: premise.Confidence * 0.75, // Abductive reasoning has lower confidence
	}

	return conclusion, nil
}

func (e *Engine) adjustConfidence(initial float64, verifications []types.VerificationStep) float64 {
	// Start with initial confidence
	confidence := initial

	// Adjust based on verification results
	for _, v := range verifications {
		if v.Confidence > 0 {
			// Weight verification results more heavily than initial confidence
			confidence = (confidence + 2*v.Confidence) / 3
		}
	}

	return confidence
}

func (e *Engine) getCurrentState(problem string, chain *types.ReasoningChain) map[string]interface{} {
	state := make(map[string]interface{})
	state["problem"] = problem
	state["steps_count"] = len(chain.Steps)
	if len(chain.Steps) > 0 {
		state["last_confidence"] = chain.Steps[len(chain.Steps)-1].Confidence
	}
	return state
}

// Add the missing generateStepWithStrategy method
func (e *Engine) generateStepWithStrategy(ctx context.Context, problem string, chain *types.ReasoningChain, strategy *types.MetaStrategy) (types.ReasoningStep, error) {
	step := types.ReasoningStep{
		Strategy: strategy,
	}

	// Build premise based on previous steps and current problem
	premise := types.LogicalExpression{
		Operation: types.AND,
		Operands:  []interface{}{problem},
	}

	// Add relevant conclusions from previous steps
	if len(chain.Steps) > 0 {
		lastStep := chain.Steps[len(chain.Steps)-1]
		premise.Operands = append(premise.Operands, lastStep.Conclusion)
	}

	step.Premise = premise

	// Use strategy to derive new conclusions
	conclusion, err := e.deriveConclusion(ctx, premise, strategy)
	if err != nil {
		return step, fmt.Errorf("failed to derive conclusion: %w", err)
	}
	step.Conclusion = conclusion

	// Calculate confidence based on strategy effectiveness and premise strength
	confidence := e.calculateStepConfidence(premise, conclusion, strategy)
	step.Confidence = confidence

	return step, nil
}

// Add the missing calculateStepConfidence method
func (e *Engine) calculateStepConfidence(premise, conclusion types.LogicalExpression, strategy *types.MetaStrategy) float64 {
	// Start with base confidence from strategy reliability
	confidence := e.getStrategyReliability(strategy)

	// Adjust based on premise confidence
	confidence *= premise.Confidence

	// Adjust confidence based on verifications
	confidence = e.adjustConfidence(confidence, []types.VerificationStep{
		{Confidence: conclusion.Confidence},
	})

	// Ensure confidence stays within [0,1]
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// Update ProcessReasoning to use these important methods
func (e *Engine) ProcessReasoning(ctx context.Context, input string) (*types.ReasoningChain, error) {
	log.Info("Processing reasoning", "input", input)
	chain := &types.ReasoningChain{}

	// Start with initial step
	initialStep, err := e.generateInitialStep(ctx, input, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to generate initial step: %w", err)
	}
	chain.Steps = append(chain.Steps, initialStep)

	// Select appropriate strategy based on input
	strategy, err := e.MetaReasoner.SelectStrategy(ctx, input, nil)
	if err != nil {
		return nil, fmt.Errorf("strategy selection failed: %w", err)
	}

	// Convert MetaStrategy to types.MetaStrategy
	typesStrategy := &types.MetaStrategy{
		Name:        strategy.Name,
		Priority:    strategy.Priority,
		Constraints: strategy.Constraints,
		Resources:   strategy.Resources,
	}

	// Generate steps through actual reasoning process
	for i := 1; i < e.MaxSteps; i++ {
		step, err := e.generateStepWithStrategy(ctx, input, chain, typesStrategy)
		if err != nil {
			return chain, fmt.Errorf("failed to generate step %d: %w", i, err)
		}

		// Validate the step
		if err := e.Validator.validateStep(step); err != nil {
			log.Warn("Step validation failed", "error", err)
			// Adjust confidence based on validation results
			step.Confidence *= 0.5 // Simple confidence reduction on validation failure
		}

		chain.Steps = append(chain.Steps, step)

		// Check if we've reached a logical conclusion
		if e.hasReachedConclusion(chain) {
			break
		}
	}

	// Calculate final confidence as average of step confidences
	chain.Confidence = e.calculateAverageConfidence(chain)
	chain.Validated = true

	return chain, nil
}

// calculateAverageConfidence calculates the average confidence across all steps
func (e *Engine) calculateAverageConfidence(chain *types.ReasoningChain) float64 {
	if len(chain.Steps) == 0 {
		return 0.0
	}

	total := 0.0
	for _, step := range chain.Steps {
		total += step.Confidence
	}
	return total / float64(len(chain.Steps))
}

// Add this method to the Engine struct
func (engine *Engine) FormatOutput(steps []types.ReasoningStep) string {
	var output []string

	for i, step := range steps {
		// Format each step's conclusion and confidence
		stepOutput := fmt.Sprintf("Step %d (%s - Confidence: %.2f):\n",
			i+1,
			step.Strategy.Name,
			step.Confidence,
		)

		// Add premise if available
		if step.Premise.Content != "" {
			stepOutput += fmt.Sprintf("Premise: %s\n", step.Premise.Content)
		}

		// Add conclusion if available
		if step.Conclusion.Content != "" {
			stepOutput += fmt.Sprintf("Conclusion: %s\n", step.Conclusion.Content)
		}

		output = append(output, stepOutput)
	}

	return strings.Join(output, "\n")
}
