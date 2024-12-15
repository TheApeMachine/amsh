// cmd/test.go
package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/errnie"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI system integration test",
	Long:  `Run a practical test of the AI system.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		errnie.Trace("%s", "test", "test")

		system := marvin.NewSystem()
		user := data.New("test", "user", "prompt", []byte("How many times do we find the letter r in the word strawberry?"))

		accumulator := twoface.NewAccumulator()
		if _, err = io.Copy(system, user); err != nil {
			return errnie.Error(err)
		}

		if _, err = io.Copy(accumulator, system); err != nil {
			return errnie.Error(err)
		}

		if _, err = io.Copy(os.Stdout, accumulator); err != nil {
			return errnie.Error(err)
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
