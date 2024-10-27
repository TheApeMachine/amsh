package reasoning

import (
	"context"
	"errors"
	"sort"
	"strings"
)

var (
	// ErrNoViableStrategy is returned when no strategy meets the requirements
	ErrNoViableStrategy = errors.New("no viable strategy found that meets constraints and resource requirements")
)

type MetaReasoner struct {
	strategies []MetaStrategy
	resources  map[string]float64
}

func NewMetaReasoner() *MetaReasoner {
	return &MetaReasoner{
		resources: make(map[string]float64),
	}
}

func (m *MetaReasoner) SelectStrategy(ctx context.Context, problem string, constraints []string) (*MetaStrategy, error) {
	// Score each strategy based on problem characteristics and constraints
	type strategyScore struct {
		strategy *MetaStrategy
		score    float64
	}

	var scores []strategyScore

	for i, strategy := range m.strategies {
		score := m.evaluateStrategy(problem, &strategy, constraints)
		scores = append(scores, strategyScore{&m.strategies[i], score})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return the highest scoring strategy that meets resource constraints
	for _, s := range scores {
		if m.canAllocateResources(s.strategy) {
			return s.strategy, nil
		}
	}

	return nil, ErrNoViableStrategy
}

func (m *MetaReasoner) evaluateStrategy(problem string, strategy *MetaStrategy, constraints []string) float64 {
	// Base score from strategy priority
	score := float64(strategy.Priority) / 10.0

	// Adjust score based on problem characteristics
	problemWords := strings.Fields(strings.ToLower(problem))
	for _, keyword := range strategy.Keywords {
		if containsWord(problemWords, keyword) {
			score += 0.15 // Higher weight for problem-specific matching
		}
	}

	// Adjust score based on constraint matching
	for _, constraint := range constraints {
		if contains(strategy.Constraints, constraint) {
			score += 0.1
		}
	}

	// Adjust score based on resource availability
	resourceScore := 0.0
	for resource, required := range strategy.Resources {
		if available, exists := m.resources[resource]; exists {
			if available >= required {
				resourceScore += 0.1
			}
		}
	}
	score += resourceScore

	// Normalize score to [0,1]
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// Helper function to check if a word exists in a slice of words
func containsWord(words []string, target string) bool {
	target = strings.ToLower(target)
	for _, word := range words {
		if strings.Contains(word, target) {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (m *MetaReasoner) canAllocateResources(strategy *MetaStrategy) bool {
	// Check if we have enough of each required resource
	for resource, required := range strategy.Resources {
		available, exists := m.resources[resource]
		if !exists || available < required {
			return false
		}
	}
	return true
}

func (m *MetaReasoner) AllocateResources(strategy *MetaStrategy) {
	// Allocate resources to the strategy
	for resource, amount := range strategy.Resources {
		m.resources[resource] -= amount
	}
}

func (m *MetaReasoner) ReleaseResources(strategy *MetaStrategy) {
	// Release resources back to the pool
	for resource, amount := range strategy.Resources {
		m.resources[resource] += amount
	}
}

func (m *MetaReasoner) RegisterStrategy(strategy MetaStrategy) {
	m.strategies = append(m.strategies, strategy)
}

func (m *MetaReasoner) InitializeResources(resources map[string]float64) {
	for k, v := range resources {
		m.resources[k] = v
	}
}

// AddStrategy is an alias for RegisterStrategy
func (m *MetaReasoner) AddStrategy(strategy MetaStrategy) {
	m.RegisterStrategy(strategy)
}
