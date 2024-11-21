// cmd/test.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai/mastercomputer"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI system integration test",
	Long:  `Run a test that demonstrates the integration between agents, communication, and VM components.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		for event := range mastercomputer.NewSystem(cmd.Context()).Input("Design a new AI research project") {
			fmt.Print(event.Content)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	os.Setenv("LOGFILE", "test.log")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}
