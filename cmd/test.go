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

	// Create a new team
	team := ai.NewTeam(toolset)

	// Add provider
	// provider, err := provider.NewRandomProvider(map[string]string{
	// 	"openai":    os.Getenv("OPENAI_API_KEY"),
	// 	"anthropic": os.Getenv("ANTHROPIC_API_KEY"),
	// 	"google":    os.Getenv("GEMINI_API_KEY"),
	// 	"cohere":    os.Getenv("COHERE_API_KEY"),
	// })
	provider := provider.NewOpenAI(
		os.Getenv("OPENAI_API_KEY"),
		"gpt-4o",
	)

	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

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

	// Set up the riddle
	riddle := "In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight."

	// Create researcher with more specific prompt
	researcher := ai.NewAgent(
		fmt.Sprintf("agent-%d", team.GetNextAgentID()),
		types.RoleResearcher,
		`You are a riddle researcher who emphasizes careful verification and challenges assumptions.
		For each finding:
		1. Question your initial assumptions
		2. Look for alternative interpretations
		3. Consider common cognitive biases
		4. Verify patterns in multiple ways
		5. Actively seek counter-evidence
		
		Remember: The most obvious pattern might be misleading.`,
		fmt.Sprintf(`Analyze this riddle with particular attention to:
		1. Multiple ways to verify patterns
		2. Potential cognitive biases in pattern recognition
		3. Alternative interpretations of "hidden" and "three"
		4. Counter-examples to test each assumption
		
		Riddle: %s`, riddle),
		toolset,
		provider,
	)

	// Create analyst with more specific prompt
	analyst := ai.NewAgent(
		fmt.Sprintf("agent-%d", team.GetNextAgentID()),
		types.RoleAnalyst,
		`You are a riddle solver who specializes in challenging assumptions.
		For each potential solution:
		1. Question your counting method
		2. Verify patterns from multiple angles
		3. Look for cognitive biases
		4. Consider alternative interpretations
		5. Actively seek evidence that would disprove your solution
		
		Remember: Simple counting tasks are often where we make the most basic mistakes.`,
		`Based on the research findings, solve the riddle by:
		1. Listing all assumptions made
		2. Verifying each pattern in multiple ways
		3. Challenging each conclusion
		4. Looking for counter-evidence
		5. Considering alternative interpretations
		
		Document your verification process and any challenges to your assumptions.`,
		toolset,
		provider,
	)

	if err := team.AddAgent(researcher); err != nil {
		return fmt.Errorf("failed to add researcher: %w", err)
	}
	if err := team.AddAgent(analyst); err != nil {
		return fmt.Errorf("failed to add analyst: %w", err)
	}

	log.Printf("Starting riddle analysis...")
	log.Printf("\nRiddle: %s\n", riddle)

	// Research phase
	researcher.SetState(types.StateWorking)
	researchResult, err := researcher.ExecuteTask()
	if err != nil {
		log.Printf("Research error: %v", err)
	}
	log.Printf("\nResearch findings:\n%s", researchResult)

	// Add initial knowledge to guide reasoning
	kb.AddFact("riddle_components", reasoning.LogicalExpression{
		Operation: reasoning.AND,
		Operands: []interface{}{
			"The answer is hidden in a fruit name",
			"Something appears three times in this fruit name",
			"The fruit is described as sweet",
		},
		Confidence: 1.0,
	})

	kb.AddFact("common_fruits", reasoning.LogicalExpression{
		Operation: reasoning.OR,
		Operands: []interface{}{
			"banana", "apple", "orange", "grape", "papaya",
			"mango", "cherry", "strawberry", "blueberry",
		},
		Confidence: 1.0,
	})

	// Convert research into logical expressions
	premise := reasoning.LogicalExpression{
		Operation:  reasoning.AND,
		Operands:   []interface{}{researchResult},
		Confidence: 0.9,
	}
	kb.AddFact("research_findings", premise)

	// Analysis phase with reasoning
	log.Printf("\nStarting reasoning process...")
	chain := &reasoning.ReasoningChain{}

	for i := 0; i < 3; i++ {
		step, err := engine.GenerateStep(cmd.Context(), researchResult, chain)
		if err != nil {
			log.Printf("Reasoning step %d error: %v", i, err)
			continue
		}

		log.Printf("\nReasoning Step %d:", i+1)
		log.Printf("Strategy: %s", step.Strategy.Name)
		log.Printf("Premise: %+v", step.Premise)
		log.Printf("Conclusion: %+v", step.Conclusion)
		log.Printf("Confidence: %.2f", step.Confidence)

		chain.Steps = append(chain.Steps, step)
	}

	// Validate reasoning
	if err := validator.ValidateChain(chain); err != nil {
		log.Printf("\nWarning: Reasoning chain validation: %v", err)
	}

	log.Printf("\nFinal Analysis:")
	log.Printf("Confidence: %.2f", chain.Confidence)
	log.Printf("Contradictions: %v", chain.Contradictions)

	// Pass to analyst for final solution
	analyst.SetState(types.StateWorking)
	if err := analyst.ReceiveMessage(fmt.Sprintf("%v", chain)); err != nil {
		log.Printf("Error passing reasoning chain to analyst: %v", err)
	}

	solution, err := analyst.ExecuteTask()
	if err != nil {
		log.Printf("Analysis error: %v", err)
	}
	log.Printf("\nProposed Solution:\n%s", solution)

	// Demonstrate team collaboration
	log.Printf("\nTeam Status Report:")
	log.Printf("Researcher state: %s", researcher.GetState())
	log.Printf("Analyst state: %s", analyst.GetState())
	log.Printf("Total messages processed: %d", team.GetMessageCount())

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
