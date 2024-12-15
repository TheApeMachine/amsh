// cmd/test.go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/data"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI system integration test",
	Long:  `Run a practical test of the AI system.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		system := marvin.NewSystem()
		user := data.New("test", "user", "prompt", []byte("How many times do we find the letter r in the word strawberry?"))

		system.Generate(user)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	os.Setenv("LOGFILE", "true")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}
