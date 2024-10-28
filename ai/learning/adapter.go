package learning

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
)

type LearningAdapter struct {
	experienceBank *ExperienceBank
}

func NewLearningAdapter() *LearningAdapter {
	return &LearningAdapter{
		experienceBank: NewExperienceBank(),
	}
}

func (la *LearningAdapter) AdaptStrategy(ctx context.Context, strategy *types.MetaStrategy, state map[string]interface{}) (*types.MetaStrategy, error) {
	// Look for patterns that match current state
	pattern, score := la.experienceBank.GetBestPattern(state, strategy.Constraints)
	if pattern != nil && score > 0.7 { // Threshold for using learned patterns
		// Adapt strategy based on learned pattern
		prvdr := provider.NewRandomProvider(map[string]string{
			"openai":    os.Getenv("OPENAI_API_KEY"),
			"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
			"google":    os.Getenv("GOOGLE_API_KEY"),
			"cohere":    os.Getenv("CLAUDE_API_KEY"),
		})
		return la.applyPattern(strategy, pattern, prvdr), nil
	}
	return strategy, nil
}

func (la *LearningAdapter) RecordStrategyExecution(strategy *types.MetaStrategy, chain *types.ReasoningChain) {
	exp := Experience{
		ID:         generateID(),
		Strategy:   strategy.Name,
		Success:    chain.Validated,
		Confidence: chain.Confidence,
		Timestamp:  time.Now(),
		Performance: PerformanceMetrics{
			SuccessRate:    calculateSuccessRate(chain),
			ConfidenceGain: calculateConfidenceGain(chain),
		},
	}

	la.experienceBank.RecordExperience(exp)
}

func (la *LearningAdapter) applyPattern(strategy *types.MetaStrategy, pattern *Pattern, provider provider.Provider) *types.MetaStrategy {
	// Create a new strategy with learned optimizations
	adaptedStrategy := *strategy // Create a copy

	// Adjust priority based on pattern reliability
	adaptedStrategy.Priority = int(float64(strategy.Priority) * pattern.Reliability)

	// Add learned keywords
	adaptedStrategy.Keywords = append(adaptedStrategy.Keywords, extractKeywords(pattern, provider)...)

	return &adaptedStrategy
}

// Helper functions
func generateID() string {
	// Use nanosecond precision for better uniqueness
	return "exp_" + time.Now().Format("20060102150405.000000000")
}

func calculateSuccessRate(chain *types.ReasoningChain) float64 {
	if len(chain.Steps) == 0 {
		return 0.0
	}

	successfulSteps := 0
	for _, step := range chain.Steps {
		if step.Confidence > 0.7 { // Consider steps with high confidence as successful
			successfulSteps++
		}
	}

	return float64(successfulSteps) / float64(len(chain.Steps))
}

func calculateConfidenceGain(chain *types.ReasoningChain) float64 {
	if len(chain.Steps) < 2 {
		return 0.0
	}

	initialConfidence := chain.Steps[0].Confidence
	finalConfidence := chain.Steps[len(chain.Steps)-1].Confidence

	return finalConfidence - initialConfidence
}

func extractKeywords(pattern *Pattern, prvdr provider.Provider) []string {
	// Create a context for the operation
	ctx := context.Background()

	// Prepare the message to send to the LLM
	message := provider.Message{
		Role:    "system",
		Content: fmt.Sprintf("Extract keywords from the following pattern: %v", pattern),
	}

	// Call the LLM to generate a response
	response, err := prvdr.GenerateSync(ctx, []provider.Message{message})
	if err != nil {
		// Handle error appropriately
		fmt.Println("Error generating keywords:", err)
		return nil
	}

	// Assume the response is a comma-separated list of keywords
	keywords := strings.Split(response, ",")

	// Clean up whitespace around keywords
	for i, keyword := range keywords {
		keywords[i] = strings.TrimSpace(keyword)
	}

	return keywords
}

func buildPatternPrompt(pattern *Pattern) string {
	var sb strings.Builder

	sb.WriteString("Pattern Analysis Request:\n\n")

	// Add state matchers
	sb.WriteString("State Conditions:\n")
	for key, value := range pattern.Trigger.StateMatchers {
		sb.WriteString("- " + key + ": " + fmt.Sprintf("%v", value) + "\n")
	}

	// Add constraints
	sb.WriteString("\nConstraints:\n")
	for _, constraint := range pattern.Trigger.Constraints {
		sb.WriteString("- " + constraint + "\n")
	}

	// Add actions
	sb.WriteString("\nActions:\n")
	for _, action := range pattern.Actions {
		sb.WriteString("- " + action.Name + "\n")
		for param, value := range action.Parameters {
			sb.WriteString("  " + param + ": " + fmt.Sprintf("%v", value) + "\n")
		}
	}

	return sb.String()
}

// Add these methods to LearningAdapter

// GetExperienceCount returns the total number of recorded experiences
func (la *LearningAdapter) GetExperienceCount() int {
	la.experienceBank.mu.RLock()
	defer la.experienceBank.mu.RUnlock()

	total := 0
	for _, experiences := range la.experienceBank.experiences {
		total += len(experiences)
	}
	return total
}

// GetAverageSuccessRate calculates the overall success rate of all experiences
func (la *LearningAdapter) GetAverageSuccessRate() float64 {
	la.experienceBank.mu.RLock()
	defer la.experienceBank.mu.RUnlock()

	var totalSuccess float64
	var totalExperiences int

	for _, experiences := range la.experienceBank.experiences {
		for _, exp := range experiences {
			if exp.Success {
				totalSuccess++
			}
			totalExperiences++
		}
	}

	if totalExperiences == 0 {
		return 0.0
	}

	return totalSuccess / float64(totalExperiences)
}

// GetPatternRecognitionRate calculates how often patterns are successfully matched
func (la *LearningAdapter) GetPatternRecognitionRate() float64 {
	la.experienceBank.mu.RLock()
	defer la.experienceBank.mu.RUnlock()

	if la.experienceBank.totalPatterns == 0 {
		return 0.0
	}

	return float64(la.experienceBank.matchedPatterns) / float64(la.experienceBank.totalPatterns)
}

// GetDetailedStats returns a comprehensive statistics report
func (la *LearningAdapter) GetDetailedStats() LearningStats {
	la.experienceBank.mu.RLock()
	defer la.experienceBank.mu.RUnlock()

	stats := LearningStats{
		TotalExperiences:  la.GetExperienceCount(),
		UniqueStrategies:  len(la.experienceBank.experiences),
		TotalPatterns:     len(la.experienceBank.patterns),
		AverageConfidence: la.calculateAverageConfidence(),
		StrategyStats:     make(map[string]StrategyStats),
		TimeBasedStats:    la.calculateTimeBasedStats(),
	}

	// Calculate per-strategy statistics
	for strategyName, experiences := range la.experienceBank.experiences {
		stats.StrategyStats[strategyName] = la.calculateStrategyStats(experiences)
	}

	return stats
}

// Supporting types for detailed statistics
type LearningStats struct {
	TotalExperiences  int
	UniqueStrategies  int
	TotalPatterns     int
	AverageConfidence float64
	StrategyStats     map[string]StrategyStats
	TimeBasedStats    TimeBasedStats
}

type StrategyStats struct {
	UseCount          int
	SuccessRate       float64
	AverageConfidence float64
	ImprovementRate   float64
}

type TimeBasedStats struct {
	LastHourSuccess float64
	LastDaySuccess  float64
	TrendSlope      float64 // Positive indicates improvement
}

// Helper methods for calculating detailed statistics
func (la *LearningAdapter) calculateAverageConfidence() float64 {
	var totalConfidence float64
	var count int

	for _, experiences := range la.experienceBank.experiences {
		for _, exp := range experiences {
			totalConfidence += exp.Confidence
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalConfidence / float64(count)
}

func (la *LearningAdapter) calculateStrategyStats(experiences []Experience) StrategyStats {
	stats := StrategyStats{
		UseCount: len(experiences),
	}

	if len(experiences) == 0 {
		return stats
	}

	var successCount, totalConfidence float64
	var confidenceChanges []float64

	for i, exp := range experiences {
		if exp.Success {
			successCount++
		}
		totalConfidence += exp.Confidence

		if i > 0 {
			change := exp.Confidence - experiences[i-1].Confidence
			confidenceChanges = append(confidenceChanges, change)
		}
	}

	stats.SuccessRate = successCount / float64(len(experiences))
	stats.AverageConfidence = totalConfidence / float64(len(experiences))

	if len(confidenceChanges) > 0 {
		var totalChange float64
		for _, change := range confidenceChanges {
			totalChange += change
		}
		stats.ImprovementRate = totalChange / float64(len(confidenceChanges))
	}

	return stats
}

func (la *LearningAdapter) calculateTimeBasedStats() TimeBasedStats {
	now := time.Now()
	hourAgo := now.Add(-time.Hour)
	dayAgo := now.Add(-24 * time.Hour)

	var hourSuccess, daySuccess float64
	var hourCount, dayCount int

	for _, experiences := range la.experienceBank.experiences {
		for _, exp := range experiences {
			if exp.Timestamp.After(hourAgo) {
				if exp.Success {
					hourSuccess++
				}
				hourCount++
			}
			if exp.Timestamp.After(dayAgo) {
				if exp.Success {
					daySuccess++
				}
				dayCount++
			}
		}
	}

	stats := TimeBasedStats{}
	if hourCount > 0 {
		stats.LastHourSuccess = hourSuccess / float64(hourCount)
	}
	if dayCount > 0 {
		stats.LastDaySuccess = daySuccess / float64(dayCount)
	}

	// Calculate trend slope using recent experiences
	stats.TrendSlope = la.calculateTrendSlope()

	return stats
}

func (la *LearningAdapter) calculateTrendSlope() float64 {
	// Simple linear regression on recent success rates
	// Positive slope indicates improvement trend
	// Implementation depends on how you want to measure trends
	return 0.0 // Placeholder
}
