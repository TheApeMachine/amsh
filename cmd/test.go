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
	"github.com/theapemachine/amsh/ai/learning" // Add this import
	"github.com/theapemachine/amsh/ai/types"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
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

// Update the passChainToAnalyst function to handle the correct Strategy type
func passChainToAnalyst(analyst *ai.Agent, chain *types.ReasoningChain, riddle, research string) error {
	// Format the chain in a way the analyst can understand
	analysis := fmt.Sprintf(`
Riddle: %s

Research Findings: %s

Reasoning Steps:
`, riddle, research)

	for i, step := range chain.Steps {
		analysis += fmt.Sprintf(`
Step %d:
- Strategy: %v
- Confidence: %.2f
`, i+1, step.Strategy, step.Confidence)
	}

	return analyst.ReceiveMessage(analysis)
}

// extractKeyTerms extracts relevant search terms from a challenge
func extractKeyTerms(challenge string) string {
	// Remove common words and punctuation
	words := strings.Fields(strings.ToLower(challenge))
	keywords := []string{}

	// Simple stopwords list
	stopwords := map[string]bool{
		"a": true, "an": true, "the": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true, "with": true,
		"by": true, "from": true, "up": true, "about": true, "into": true,
		"over": true, "after": true,
	}

	// Extract meaningful keywords
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?")
		if len(word) > 2 && !stopwords[word] {
			keywords = append(keywords, word)
		}
	}

	return strings.Join(keywords, " ")
}

// identifyProblemType determines the type of challenge
func identifyProblemType(challenge string) string {
	challenge = strings.ToLower(challenge)

	switch {
	case strings.Contains(challenge, "sequence") || strings.Contains(challenge, "number"):
		return "mathematical"
	case strings.Contains(challenge, "transform") || strings.Contains(challenge, "word"):
		return "linguistic"
	case strings.Contains(challenge, "riddle"):
		return "riddle"
	default:
		return "logical"
	}
}

// assessComplexity evaluates the complexity of the challenge
func assessComplexity(challenge string) float64 {
	// Base complexity
	complexity := 1.0

	// Factors that increase complexity
	if strings.Contains(strings.ToLower(challenge), "explain") {
		complexity += 0.2
	}
	if strings.Count(challenge, "?") > 1 {
		complexity += 0.3
	}
	if len(strings.Split(challenge, " ")) > 20 {
		complexity += 0.4
	}

	// Cap complexity between 0 and 1
	if complexity > 1.0 {
		complexity = 1.0
	}

	return complexity
}
