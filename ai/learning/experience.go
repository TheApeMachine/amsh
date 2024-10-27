package learning

import (
	"sync"
	"time"
)

// Experience represents a single learning instance
type Experience struct {
	ID             string
	Strategy       string
	InitialState   map[string]interface{}
	Actions        []Action
	FinalState     map[string]interface{}
	Success        bool
	Confidence     float64
	Timestamp      time.Time
	Performance    PerformanceMetrics
	PatternMatched bool // Add this field
}

type Action struct {
	Name       string
	Parameters map[string]interface{}
	Result     string
	Duration   time.Duration
}

type PerformanceMetrics struct {
	ExecutionTime  time.Duration
	ResourceUsage  map[string]float64
	SuccessRate    float64
	ConfidenceGain float64
	ErrorFrequency int
}

// ExperienceBank stores and manages learning experiences
type ExperienceBank struct {
	experiences     map[string][]Experience
	patterns        map[string]Pattern
	mu              sync.RWMutex
	totalPatterns   int
	matchedPatterns int
}

type Pattern struct {
	Trigger     Condition
	Actions     []Action
	Reliability float64
	UseCount    int
}

type Condition struct {
	StateMatchers map[string]interface{}
	Constraints   []string
}

func NewExperienceBank() *ExperienceBank {
	return &ExperienceBank{
		experiences: make(map[string][]Experience),
		patterns:    make(map[string]Pattern),
	}
}

func (eb *ExperienceBank) RecordExperience(exp Experience) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Store the experience
	eb.experiences[exp.Strategy] = append(eb.experiences[exp.Strategy], exp)

	// Update patterns based on new experience
	eb.updatePatterns(exp)
}

func (eb *ExperienceBank) updatePatterns(exp Experience) {
	// Extract pattern from experience
	pattern := eb.extractPattern(exp)

	// Update existing pattern or create new one
	if existing, exists := eb.patterns[pattern.Trigger.String()]; exists {
		// Update reliability based on success
		totalCount := float64(existing.UseCount + 1)
		if exp.Success {
			existing.Reliability = (existing.Reliability*float64(existing.UseCount) + 1.0) / totalCount
		} else {
			existing.Reliability = (existing.Reliability * float64(existing.UseCount)) / totalCount
		}
		existing.UseCount++
		eb.patterns[pattern.Trigger.String()] = existing
	} else {
		// Create new pattern with initial reliability
		initialReliability := 0.0
		if exp.Success {
			initialReliability = 1.0
		}

		eb.patterns[pattern.Trigger.String()] = Pattern{
			Trigger:     pattern.Trigger,
			Actions:     exp.Actions,
			Reliability: initialReliability,
			UseCount:    1,
		}
	}
}

func (eb *ExperienceBank) GetBestPattern(state map[string]interface{}, constraints []string) (*Pattern, float64) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var bestPattern *Pattern
	var bestScore float64

	for _, pattern := range eb.patterns {
		if score := pattern.matchScore(state, constraints); score > bestScore {
			bestScore = score
			bestPattern = &pattern
		}
	}

	return bestPattern, bestScore
}

func (p *Pattern) matchScore(state map[string]interface{}, constraints []string) float64 {
	// Base score from reliability
	score := p.Reliability

	// Adjust based on state matching
	matchCount := 0
	for key, value := range p.Trigger.StateMatchers {
		if stateValue, exists := state[key]; exists && stateValue == value {
			matchCount++
		}
	}
	stateScore := float64(matchCount) / float64(len(p.Trigger.StateMatchers))

	// Adjust based on constraints matching
	constraintScore := 0.0
	for _, constraint := range constraints {
		for _, patternConstraint := range p.Trigger.Constraints {
			if constraint == patternConstraint {
				constraintScore += 1.0
				break
			}
		}
	}
	constraintScore /= float64(len(constraints))

	// Combine scores with weights
	return score*0.4 + stateScore*0.4 + constraintScore*0.2
}

func (c *Condition) String() string {
	// Implement a stable string representation of the condition
	// This is important for using conditions as map keys
	// TODO: Implement proper string serialization
	return "condition_string"
}

func (eb *ExperienceBank) extractPattern(exp Experience) Pattern {
	return Pattern{
		Trigger: Condition{
			StateMatchers: exp.InitialState,
			Constraints:   []string{}, // Extract constraints from experience
		},
		Actions: exp.Actions,
	}
}
