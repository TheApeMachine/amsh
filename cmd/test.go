// cmd/test.go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/ai/marvin/interaction"
	"github.com/theapemachine/amsh/ai/provider"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI system integration test",
	Long:  `Run a test that demonstrates the integration between agents, communication, and VM components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Create agents with different roles
		planner := marvin.NewAgent(ctx, "planner")
		critic := marvin.NewAgent(ctx, "critic")
		executor := marvin.NewAgent(ctx, "executor")

		// Set up the test prompt
		testPrompt := `Let's solve this problem step by step:
How many times do we find the letter 'r' in the word 'strawberry'?
Please explain your reasoning.`

		// Set prompts for all agents
		planner.SetUserPrompt(testPrompt)
		critic.SetUserPrompt("Analyze the solution and check for errors.")
		executor.SetUserPrompt("Execute the final solution and verify the result.")

		// Create interactions
		memoryInteraction := interaction.NewMemoryInteraction(planner)
		discussionInteraction := interaction.NewDiscussionInteraction([]*marvin.Agent{planner, critic})
		toolInteraction := interaction.NewToolInteraction(executor)

		// Configure tool interaction with a simple counter tool
		toolInteraction.AddTool("counter", map[string]interface{}{
			"count": func(s, char string) int {
				count := 0
				for _, c := range s {
					if string(c) == char {
						count++
					}
				}
				return count
			},
		})

		// Create processor with all interactions
		processor := marvin.NewProcessor(ctx, memoryInteraction, discussionInteraction, toolInteraction)

		// Configure interactions
		configs := map[string]map[string]interface{}{
			"discussion": {
				"rounds":    3,
				"consensus": 0.8,
			},
		}
		if err := processor.Configure(configs); err != nil {
			return err
		}

		// Process and display results
		fmt.Println("Starting AI system integration test...")
		fmt.Println("Problem:", testPrompt)
		fmt.Println("\nProcessing...")

		for event := range processor.Process() {
			switch event.Type {
			case provider.EventToken:
				fmt.Print(event.Content)
			case provider.EventToolCall:
				fmt.Printf("\nTool Result: %s\n", event.Content)
			case provider.EventError:
				fmt.Printf("\nError: %v\n", event.Error)
			case provider.EventDone:
				fmt.Println("\nProcessing complete.")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	os.Setenv("LOGFILE", "true")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}
