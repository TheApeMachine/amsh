package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3/client"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/learning"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/reasoning"
	"github.com/theapemachine/amsh/ai/types"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	toolset, err := ai.NewToolset()
	if err != nil {
		return fmt.Errorf("failed to create toolset: %w", err)
	}

	// Create a new team and get agents
	team := ai.NewTeam(toolset)
	researcher := team.GetResearcher()
	analyst := team.GetAnalyst()

	// Initialize the LLM provider
	llm := provider.NewOpenAI(
		os.Getenv("OPENAI_API_KEY"),
		"gpt-4",
	)
	team.SetProvider(llm)

	// Initialize reasoning system
	kb := reasoning.NewKnowledgeBase()
	validator := reasoning.NewValidator(kb)
	metaReasoner := reasoning.NewMetaReasoner()

	// Initialize resources
	metaReasoner.InitializeResources(map[string]float64{
		"cpu":    1.0,
		"memory": 1.0,
		"time":   1.0,
	})

	// Register specialized strategies for riddle solving
	metaReasoner.RegisterStrategy(reasoning.MetaStrategy{
		Name:     "pattern_analysis",
		Priority: 8,
		Resources: map[string]float64{
			"cpu":    0.5,
			"memory": 0.3,
		},
		Constraints: []string{"pattern_matching", "word_analysis"},
	})

	metaReasoner.RegisterStrategy(reasoning.MetaStrategy{
		Name:     "word_decomposition",
		Priority: 7,
		Resources: map[string]float64{
			"cpu":    0.4,
			"memory": 0.2,
		},
		Constraints: []string{"linguistic_analysis", "creative"},
	})

	metaReasoner.RegisterStrategy(reasoning.MetaStrategy{
		Name:     "semantic_connection",
		Priority: 6,
		Resources: map[string]float64{
			"cpu":    0.3,
			"memory": 0.2,
		},
		Constraints: []string{"meaning_analysis", "context_aware"},
	})

	// Create engine using constructor
	engine := reasoning.NewEngine(validator, metaReasoner)

	// Create initial strategy as types.MetaStrategy
	baseStrategy := &types.MetaStrategy{
		Name:     "pattern_analysis",
		Priority: 8,
		Resources: map[string]float64{
			"cpu":    0.5,
			"memory": 0.3,
		},
		Constraints: []string{"pattern_matching", "word_analysis"},
	}

	// Initialize learning system
	learningAdapter := learning.NewLearningAdapter()

	// Create a slice of test riddles to demonstrate learning
	riddles := []string{
		"In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight.",
		"Round as a button, deep as a cup, all the king's horses can't fill it up.",
		"What has keys but no locks, space but no room, and you can enter but not go in?",
	}

	log.Printf("Starting adaptive learning test with multiple riddles...\n")

	for i, riddle := range riddles {
		log.Printf("\n=== Riddle %d ===\n%s\n", i+1, riddle)

		state := map[string]interface{}{
			"problem_type": "riddle",
			"complexity":   "medium",
			"attempt":      i + 1,
		}

		adaptedStrategy, err := learningAdapter.AdaptStrategy(cmd.Context(), baseStrategy, state)
		if err != nil {
			log.Printf("Strategy adaptation warning: %v", err)
		}

		// Convert types.MetaStrategy to reasoning.MetaStrategy for the meta reasoner
		reasoningStrategy := reasoning.MetaStrategy{
			Name:        adaptedStrategy.Name,
			Priority:    adaptedStrategy.Priority,
			Resources:   adaptedStrategy.Resources,
			Constraints: adaptedStrategy.Constraints,
		}
		metaReasoner.RegisterStrategy(reasoningStrategy)

		// Research phase
		researcher.SetState(types.StateWorking)
		researchResult, err := researcher.ExecuteTask()
		if err != nil {
			log.Printf("Research error: %v", err)
		}
		log.Printf("\nResearch findings:\n%s", researchResult)

		// Reasoning phase with type conversion
		typesChain := &types.ReasoningChain{}
		for step := 0; step < 3; step++ {
			reasoningStep, err := engine.GenerateStep(cmd.Context(), researchResult, typesChain)
			if err != nil {
				log.Printf("Reasoning step %d error: %v", step, err)
				continue
			}

			typesChain.Steps = append(typesChain.Steps, reasoningStep)
			log.Printf("\nReasoning Step %d:", step+1)
			log.Printf("Strategy: %s", reasoningStep.Strategy.Name)
			log.Printf("Confidence: %.2f", reasoningStep.Confidence)
		}

		// Analysis phase
		analyst.SetState(types.StateWorking)
		if err := analyst.ReceiveMessage(fmt.Sprintf("%v", typesChain)); err != nil {
			log.Printf("Error passing reasoning chain to analyst: %v", err)
		}

		solution, err := analyst.ExecuteTask()
		if err != nil {
			log.Printf("Analysis error: %v", err)
		}
		log.Printf("\nProposed Solution:\n%s", solution)

		// Record the experience
		learningAdapter.RecordStrategyExecution(adaptedStrategy, typesChain)

		// Display learning metrics
		log.Printf("\n=== Learning Metrics ===")
		log.Printf("Success Rate: %.2f", calculateSuccessRate(typesChain))
		log.Printf("Confidence Gain: %.2f", calculateConfidenceGain(typesChain))
		log.Printf("Strategy Reliability: %.2f", float64(adaptedStrategy.Priority)/float64(baseStrategy.Priority))
	}

	// Display final learning statistics
	displayLearningStats(learningAdapter)

	// Clean up
	team.Shutdown()
	log.Printf("Team shutdown completed")

	if err := deleteQdrantCollections(); err != nil {
		log.Printf("Error deleting Qdrant collections: %v", err)
	}

	return nil
}

func deleteQdrantCollections() error {
	// List collections
	resp, err := client.Get("http://localhost:6333/collections")
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	var listResp struct {
		Result struct {
			Collections []struct {
				Name string `json:"name"`
			} `json:"collections"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp.Body(), &listResp); err != nil {
		return fmt.Errorf("failed to unmarshal collection list: %w", err)
	}

	// Delete collections
	for _, collection := range listResp.Result.Collections {
		if strings.Contains(collection.Name, "marvin") || strings.Contains(collection.Name, "hive") {
			log.Printf("Skipping deletion of protected collection: %s", collection.Name)
			continue
		}

		resp, err := client.Delete(fmt.Sprintf("http://localhost:6333/collections/%s", collection.Name))
		if err != nil {
			log.Printf("Failed to delete collection %s: %v", collection.Name, err)
			continue
		}

		if resp.StatusCode() != http.StatusOK {
			log.Printf("Failed to delete collection %s: status code %d", collection.Name, resp.StatusCode())
		} else {
			log.Printf("Successfully deleted collection: %s", collection.Name)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
	os.Setenv("LOGFILE", "test.log")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}

func calculateSuccessRate(chain *types.ReasoningChain) float64 {
	if len(chain.Steps) == 0 {
		return 0.0
	}

	successfulSteps := 0
	for _, step := range chain.Steps {
		if step.Confidence > 0.7 {
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

func displayLearningStats(adapter *learning.LearningAdapter) {
	stats := adapter.GetDetailedStats()

	log.Printf("=== Learning System Statistics ===")
	log.Printf("Total Experiences: %d", stats.TotalExperiences)
	log.Printf("Unique Strategies: %d", stats.UniqueStrategies)
	log.Printf("Total Patterns: %d", stats.TotalPatterns)
	log.Printf("Average Confidence: %.2f", stats.AverageConfidence)

	log.Printf("\nPer-Strategy Performance:")
	for strategy, stratStats := range stats.StrategyStats {
		log.Printf("\n  Strategy: %s", strategy)
		log.Printf("    Use Count: %d", stratStats.UseCount)
		log.Printf("    Success Rate: %.2f", stratStats.SuccessRate)
		log.Printf("    Average Confidence: %.2f", stratStats.AverageConfidence)
		log.Printf("    Improvement Rate: %.2f", stratStats.ImprovementRate)
	}

	log.Printf("\nTime-Based Analysis:")
	log.Printf("  Last Hour Success Rate: %.2f", stats.TimeBasedStats.LastHourSuccess)
	log.Printf("  Last Day Success Rate: %.2f", stats.TimeBasedStats.LastDaySuccess)
	log.Printf("  Learning Trend: %.2f", stats.TimeBasedStats.TrendSlope)
}
