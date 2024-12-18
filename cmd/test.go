// cmd/test.go
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/marvin"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/errnie"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI system integration test",
	Long:  `Run a practical test of the AI system.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		os.Setenv("LOGFILE", "true")
		errnie.InitLogger()

		agent := marvin.NewAgent(
			context.Background(),
			"test",
			"prompt",
			data.New("test", "system", "prompt", []byte(
				viper.GetViper().GetString("ai.setup.marvin.templates.system"),
			)),
		)

		user := data.New(
			"test",
			"user",
			"prompt",
			[]byte("Develop a simple web application that allows users to upload files and download them."),
		)

		sidekick := marvin.NewAgent(
			context.Background(),
			"sidekick",
			"prompt",
			data.New("test", "system", "prompt", []byte(
				viper.GetViper().GetString("ai.setups.marvin.tools.docker.description")),
			),
		)
		sidekick.AddTools(tools.NewEnvironment())

		agent.AddSidekick("developer", sidekick)

		for artifact := range agent.Generate(user) {
			fmt.Print(string(artifact.Peek("payload")))
		}

		fmt.Println("")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	os.Setenv("LOGFILE", "true")
	os.Setenv("QDRANT_URL", "http://localhost:6333")
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}
