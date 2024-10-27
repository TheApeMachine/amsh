package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/client"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/codegen"
	"github.com/theapemachine/amsh/ai/learning"
	"github.com/theapemachine/amsh/ai/planning" // Add this import
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/reasoning"
	"github.com/theapemachine/amsh/ai/tools"
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
	if researcher == nil {
		return fmt.Errorf("failed to get researcher agent")
	}

	analyst := team.GetAnalyst()
	if analyst == nil {
		return fmt.Errorf("failed to get analyst agent")
	}

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

	// Register strategies for different types of challenges
	metaReasoner.RegisterStrategy(reasoning.MetaStrategy{
		Name:     "numerical_analysis",
		Priority: 8,
		Resources: map[string]float64{
			"cpu":    0.6,
			"memory": 0.4,
		},
		Constraints: []string{"pattern_matching", "mathematical_reasoning"},
	})

	metaReasoner.RegisterStrategy(reasoning.MetaStrategy{
		Name:     "linguistic_transformation",
		Priority: 7,
		Resources: map[string]float64{
			"cpu":    0.5,
			"memory": 0.3,
		},
		Constraints: []string{"word_analysis", "dictionary_lookup"},
	})

	metaReasoner.RegisterStrategy(reasoning.MetaStrategy{
		Name:     "logical_deduction",
		Priority: 8,
		Resources: map[string]float64{
			"cpu":    0.7,
			"memory": 0.5,
		},
		Constraints: []string{"problem_solving", "step_by_step_analysis"},
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
	challenges := []string{
		// Original fruit riddle that requires actual reasoning
		"In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight.",

		// New pattern-based challenge
		"Create a sequence where each number is the sum of the digits of the previous number multiplied by 2. Starting with 14, what are the next 5 numbers and what pattern emerges?",

		// Logic puzzle that requires research and reasoning
		"A new species of flower was discovered that doubles its petals every day. If it takes 8 days to fill a garden, how many days would it take if we started with two flowers instead of one? Explain your reasoning.",

		// Word transformation challenge
		"Transform the word 'CODE' into 'DATA' by changing one letter at a time, making valid English words at each step. What's the shortest path and why?",
	}

	log.Printf("Starting adaptive learning test with multiple riddles...\n")

	// Initialize the planner
	planner := planning.NewPlanner()

	// Create an example plan for the AI pipeline
	planReq := planning.CreatePlanRequest{
		Name:        "AI Pipeline Execution",
		Description: "Coordinated execution of AI agents for problem solving",
		EndTime:     time.Now().Add(1 * time.Hour),
		Goals: []planning.CreateGoalRequest{
			{
				Name:        "Problem Analysis",
				Description: "Analyze and understand the given problem",
				Priority:    1,
				Deadline:    time.Now().Add(15 * time.Minute),
				Objectives: []planning.CreateObjectiveRequest{
					{
						Name:        "Research Phase",
						Description: "Gather relevant information",
						Deadline:    time.Now().Add(5 * time.Minute),
					},
					{
						Name:        "Pattern Analysis",
						Description: "Identify patterns and relationships",
						Deadline:    time.Now().Add(10 * time.Minute),
					},
				},
			},
			{
				Name:        "Solution Generation",
				Description: "Generate and validate solutions",
				Priority:    2,
				Deadline:    time.Now().Add(45 * time.Minute),
				Objectives: []planning.CreateObjectiveRequest{
					{
						Name:        "Strategy Application",
						Description: "Apply reasoning strategies",
						Deadline:    time.Now().Add(30 * time.Minute),
					},
					{
						Name:        "Solution Validation",
						Description: "Validate proposed solutions",
						Deadline:    time.Now().Add(40 * time.Minute),
					},
				},
			},
		},
	}

	plan, err := planner.CreatePlan(cmd.Context(), planReq)
	if err != nil {
		return fmt.Errorf("failed to create execution plan: %w", err)
	}

	log.Printf("\n=== Execution Plan Created ===")
	log.Printf("Plan ID: %s", plan.ID)
	log.Printf("Plan Name: %s", plan.Name)
	log.Printf("Status: %s", plan.Status)

	// Initialize code generator
	generator, err := codegen.NewGenerator()
	if err != nil {
		return fmt.Errorf("failed to create code generator: %w", err)
	}

	// Initialize browser tool
	browser := tools.NewBrowser()
	if err := browser.StartSession(); err != nil {
		return fmt.Errorf("failed to start browser session: %w", err)
	}
	defer browser.Instance.Close() // Changed from instance to Instance

	// Add tools to toolset
	toolset.AddTool("codegen", generator)
	toolset.AddTool("browser", browser)

	// Update the riddles loop to include code generation and browser research
	for i, challenge := range challenges {
		log.Printf("\n=== Challenge %d ===\n%s\n", i+1, challenge)

		// Use browser to research relevant concepts
		searchTerms := extractKeyTerms(challenge) // We'd need to implement this
		browserResult, err := browser.Execute(cmd.Context(), map[string]interface{}{
			"url":      fmt.Sprintf("https://www.google.com/search?q=%s", url.QueryEscape(searchTerms)),
			"selector": "div.g",
			"timeout":  5.0,
		})
		if err != nil {
			log.Printf("Browser research error: %v", err)
		} else {
			log.Printf("\nRelevant Research:\n%s", browserResult)
		}

		// Update the state to reflect the type of challenge
		state := map[string]interface{}{
			"problem_type": identifyProblemType(challenge), // We'd need to implement this
			"complexity":   assessComplexity(challenge),    // We'd need to implement this
			"attempt":      i + 1,
		}

		adaptedStrategy, err := learningAdapter.AdaptStrategy(cmd.Context(), baseStrategy, state)
		if err != nil {
			log.Printf("Strategy adaptation warning: %v", err)
			continue
		}

		// Send riddle to researcher
		if err := researcher.ReceiveMessage(challenge); err != nil {
			log.Printf("Error sending riddle to researcher: %v", err)
			continue
		}

		// Get research findings
		findings, err := researcher.ExecuteTask()
		if err != nil {
			log.Printf("Research error: %v", err)
			continue
		}
		log.Printf("\nResearch findings:\n%s", findings)

		// Process reasoning chain
		chain, err := engine.ProcessReasoning(cmd.Context(), challenge)
		if err != nil {
			log.Printf("Error processing reasoning chain: %v", err)
			continue
		}
		if chain == nil {
			log.Printf("Error: nil reasoning chain")
			continue
		}

		// Convert reasoning chain for analyst
		typesChain := &types.ReasoningChain{
			Steps: make([]types.ReasoningStep, len(chain.Steps)),
		}

		for i, step := range chain.Steps {
			typesChain.Steps[i] = types.ReasoningStep{
				Strategy:   step.Strategy,
				Confidence: step.Confidence,
			}
			log.Printf("\nReasoning Step %d:", i+1)
			log.Printf("Strategy: %v", step.Strategy)
			log.Printf("Confidence: %.2f", step.Confidence)
		}

		// Pass research findings along with the chain
		if err := passChainToAnalyst(analyst, typesChain, challenge, findings); err != nil {
			log.Printf("Error passing reasoning chain to analyst: %v", err)
			continue
		}

		solution, err := analyst.ExecuteTask()
		if err != nil {
			log.Printf("Analysis error: %v", err)
			continue
		}
		log.Printf("\nProposed Solution:\n%s", solution)

		// Record the experience
		learningAdapter.RecordStrategyExecution(adaptedStrategy, typesChain)

		// Display learning metrics
		log.Printf("\n=== Learning Metrics ===")
		log.Printf("Success Rate: %.2f", calculateSuccessRate(typesChain))
		log.Printf("Confidence Gain: %.2f", calculateConfidenceGain(typesChain))
		log.Printf("Strategy Reliability: %.2f", float64(adaptedStrategy.Priority)/float64(baseStrategy.Priority))

		// Update plan with research completion
		researchUpdate := planning.PlanUpdates{
			TaskUpdates: []planning.TaskUpdate{
				{
					TaskID:   fmt.Sprintf("research_%d", i),
					Progress: 1.0,
					Status:   planning.TaskStatusComplete,
				},
			},
		}

		if err := planner.UpdatePlan(cmd.Context(), plan.ID, researchUpdate); err != nil {
			log.Printf("Warning: failed to update plan for research completion: %v", err)
		}

		// Update plan with analysis phase
		analysisUpdate := planning.PlanUpdates{
			TaskUpdates: []planning.TaskUpdate{
				{
					TaskID:   fmt.Sprintf("analysis_%d", i),
					Progress: 0.0,
					Status:   planning.TaskStatusActive,
				},
			},
		}

		if err := planner.UpdatePlan(cmd.Context(), plan.ID, analysisUpdate); err != nil {
			log.Printf("Warning: failed to update plan for analysis phase: %v", err)
		}

		// Add browser-based research capability
		if err := researcher.ReceiveMessage(fmt.Sprintf("Use the browser to research: %s", challenge)); err != nil {
			log.Printf("Error sending research request: %v", err)
			continue
		}

		// Generate solution code
		if solution != "" {
			codeResult, err := generator.Execute(cmd.Context(), map[string]interface{}{
				"language": "Go",
				"name":     fmt.Sprintf("riddle_solver_%d", i),
				"description": fmt.Sprintf(`
					// RiddleSolver implements a solution for the riddle:
					// %s
					//
					// Solution: %s
				`, challenge, solution),
				"tests": true,
			})
			if err != nil {
				log.Printf("Code generation error: %v", err)
			} else {
				log.Printf("\nGenerated Solution Code:\n%s", codeResult)
			}
		}
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
