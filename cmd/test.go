package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/mastercomputer"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := data.New(
			"user",
			"prompt",
			"task",
			[]byte{},
		)

		prompt.Poke("system", "You are part of an advanced AI system, called The Ape Machine.")
		prompt.Poke("guidelines", "If at some point you do not know what to do, you should just say so and ask for help. Never make up an answer, or try to guess at an answer. You have a fully functional Linux environment tool.")
		prompt.Poke("toolset", "system")
		prompt.Poke("user", "Discover your current environment.")

		systems := []*mastercomputer.Worker{
			mastercomputer.NewWorker().Initialize(cmd.Context(), prompt),
		}

		for _, system := range systems {
			system.Run(context.Background())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
