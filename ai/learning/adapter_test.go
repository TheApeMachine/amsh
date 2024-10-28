package learning

import (
	"context"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
)

func TestLearningAdapter(t *testing.T) {
	Convey("Given a learning adapter", t, func() {
		adapter := NewLearningAdapter()

		Convey("When adapting a strategy", func() {
			strategy := &types.MetaStrategy{
				Name:     "test_strategy",
				Priority: 5,
				Constraints: []string{
					"time_critical",
					"pattern_matching",
				},
				Keywords: []string{"initial_keyword"},
			}

			state := map[string]interface{}{
				"problem_type": "riddle",
				"complexity":   "high",
			}

			adaptedStrategy, err := adapter.AdaptStrategy(context.Background(), strategy, state)

			Convey("Then it should return the strategy without error", func() {
				So(err, ShouldBeNil)
				So(adaptedStrategy, ShouldNotBeNil)
				So(adaptedStrategy.Name, ShouldEqual, strategy.Name)
			})
		})

		Convey("When recording strategy execution", func() {
			strategy := &types.MetaStrategy{
				Name:     "test_strategy",
				Priority: 5,
			}

			chain := &types.ReasoningChain{
				Steps: []types.ReasoningStep{
					{
						Confidence: 0.8,
					},
					{
						Confidence: 0.9,
					},
				},
				Validated:  true,
				Confidence: 0.85,
			}

			Convey("Then it should record the experience", func() {
				adapter.RecordStrategyExecution(strategy, chain)
				// Verify the experience was recorded by checking the experience bank
				experiences := adapter.experienceBank.experiences[strategy.Name]
				So(len(experiences), ShouldEqual, 1)
				So(experiences[0].Success, ShouldBeTrue)
				So(experiences[0].Confidence, ShouldEqual, 0.85)
			})
		})

		Convey("When extracting keywords from a pattern", func() {
			// Skip if no API keys are set
			if os.Getenv("OPENAI_API_KEY") == "" &&
				os.Getenv("ANTHROPIC_API_KEY") == "" &&
				os.Getenv("GOOGLE_API_KEY") == "" &&
				os.Getenv("COHERE_API_KEY") == "" {
				SkipConvey("Skipping keyword extraction test - no API keys set", func() {})
				return
			}

			pattern := &Pattern{
				Trigger: Condition{
					StateMatchers: map[string]interface{}{
						"problem_type": "riddle",
						"complexity":   "high",
					},
					Constraints: []string{"time_critical", "pattern_matching"},
				},
				Actions: []Action{
					{
						Name: "analyze_patterns",
						Parameters: map[string]interface{}{
							"method": "frequency_analysis",
						},
					},
				},
				Reliability: 0.8,
			}

			prvdr := provider.NewRandomProvider(map[string]string{
				"openai":    os.Getenv("OPENAI_API_KEY"),
				"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
				"google":    os.Getenv("GOOGLE_API_KEY"),
				"cohere":    os.Getenv("CLAUDE_API_KEY"),
			})
			keywords := extractKeywords(pattern, prvdr)

			Convey("Then it should return relevant keywords", func() {
				So(keywords, ShouldNotBeEmpty)
				// Check that common expected keywords are present
				foundRelevantKeyword := false
				for _, keyword := range keywords {
					if keyword == "riddle" || keyword == "pattern" ||
						keyword == "analysis" || keyword == "time_critical" {
						foundRelevantKeyword = true
						break
					}
				}
				So(foundRelevantKeyword, ShouldBeTrue)
			})
		})

		Convey("When calculating success rate", func() {
			chain := &types.ReasoningChain{
				Steps: []types.ReasoningStep{
					{Confidence: 0.8}, // Success
					{Confidence: 0.6}, // Failure
					{Confidence: 0.9}, // Success
				},
			}

			rate := calculateSuccessRate(chain)

			Convey("Then it should return the correct rate", func() {
				So(rate, ShouldEqual, 2.0/3.0)
			})
		})

		Convey("When calculating confidence gain", func() {
			chain := &types.ReasoningChain{
				Steps: []types.ReasoningStep{
					{Confidence: 0.5},
					{Confidence: 0.8},
				},
			}

			gain := calculateConfidenceGain(chain)

			Convey("Then it should return the correct gain", func() {
				// Use ShouldAlmostEqual for floating point comparisons
				So(gain, ShouldAlmostEqual, 0.3, 0.000001)
			})
		})

		Convey("When generating an ID", func() {
			// Generate IDs with nanosecond precision to ensure uniqueness
			id1 := generateID()
			id2 := generateID()

			Convey("Then it should generate unique IDs", func() {
				So(id1, ShouldNotEqual, id2)
				So(id1, ShouldStartWith, "exp_")
				So(id2, ShouldStartWith, "exp_")
			})
		})

		Convey("When building a pattern prompt", func() {
			pattern := &Pattern{
				Trigger: Condition{
					StateMatchers: map[string]interface{}{
						"test_state": "test_value",
					},
					Constraints: []string{"test_constraint"},
				},
				Actions: []Action{
					{
						Name: "test_action",
						Parameters: map[string]interface{}{
							"param": "value",
						},
					},
				},
			}

			prompt := buildPatternPrompt(pattern)

			Convey("Then it should build a well-formatted prompt", func() {
				So(prompt, ShouldContainSubstring, "Pattern Analysis Request")
				So(prompt, ShouldContainSubstring, "State Conditions")
				So(prompt, ShouldContainSubstring, "test_state: test_value")
				So(prompt, ShouldContainSubstring, "Constraints")
				So(prompt, ShouldContainSubstring, "test_constraint")
				So(prompt, ShouldContainSubstring, "Actions")
				So(prompt, ShouldContainSubstring, "test_action")
				So(prompt, ShouldContainSubstring, "param: value")
			})
		})

		Convey("When applying a pattern to a strategy", func() {
			strategy := &types.MetaStrategy{
				Name:     "test_strategy",
				Priority: 10,
				Keywords: []string{"initial_keyword"},
			}

			pattern := &Pattern{
				Trigger: Condition{
					StateMatchers: map[string]interface{}{
						"problem_type": "riddle",
						"complexity":   "high",
					},
					Constraints: []string{"time_critical"},
				},
				Actions: []Action{
					{
						Name: "analyze_patterns",
						Parameters: map[string]interface{}{
							"method": "frequency_analysis",
						},
					},
				},
				Reliability: 0.8,
			}

			prvdr := provider.NewRandomProvider(map[string]string{
				"openai":    os.Getenv("OPENAI_API_KEY"),
				"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
				"google":    os.Getenv("GOOGLE_API_KEY"),
				"cohere":    os.Getenv("CLAUDE_API_KEY"),
			})
			adaptedStrategy := adapter.applyPattern(strategy, pattern, prvdr)

			Convey("Then it should create an adapted strategy", func() {
				So(adaptedStrategy, ShouldNotBeNil)
				So(adaptedStrategy.Name, ShouldEqual, strategy.Name)
				// Priority should be adjusted based on pattern reliability
				So(adaptedStrategy.Priority, ShouldEqual, int(float64(strategy.Priority)*pattern.Reliability))
				// Should preserve original keywords
				So(adaptedStrategy.Keywords, ShouldContain, "initial_keyword")
				// Should not modify original strategy
				So(strategy.Priority, ShouldEqual, 10)
			})
		})
	})
}
