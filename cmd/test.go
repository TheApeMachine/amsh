// cmd/test.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai/marvin"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI system integration test",
	Long:  `Run a test that demonstrates the integration between agents, communication, and VM components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		for event := range marvin.NewSystem().Process("How many times do we find the letter r in the word strawberry?") {
			fmt.Print(event.Content)
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
