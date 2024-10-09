package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/mastercomputer"
)

var smallTest = []string{
	"How many time do we find the letter r in the word strawberry?",
}

var largeTest = []string{
	"How many time do we find the letter r in the word strawberry?",
	"What is the capital of the moon?",
	"Solve the riddle: In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight.",
	"Question: Suppose you’re on a game show, and you’re given the choice of three doors: Behind one door is a gold bar; behind the others, rotten vegetables. You pick a door, say No.1, and the host asks you “Do you want to pick door No.2 instead?” Is it to your advantage to switch your choice?",
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		systems := []*mastercomputer.Worker{
			mastercomputer.NewWorker().Initialize(),
		}

		prompt := data.New(
			"user",
			"prompt",
			"discussion",
			[]byte("Solve the riddle: In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight."),
		)

		prompt.SetAttrs(map[string]string{
			"system": "You are part of an advanced AI system, called The Ape Machine.",
			"user":   "If at some point you do not know what to do, you should just say so and ask for help. Never make up an answer, or try to guess at an answer.",
		})

		for _, system := range systems {
			prompt = system.Run(context.Background(), prompt)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
