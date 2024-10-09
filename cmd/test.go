package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		system := mastercomputer.NewWorker()
		system.Initialize()
		system.Run(context.Background(), "test", map[string]any{
			"system":  "The Ape Machine is an AI-powered Operating System, designed to handle any task, environment, or user.",
			"user":    "I need you to build the workers you think you need, and then let me know when you're ready.",
			"toolset": "test",
		})

		for {
			// Get user input from the terminal
			fmt.Print("> ")
			reader := bufio.NewReader(os.Stdin)
			prompt, _ := reader.ReadString('\n')
			system.I <- prompt
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
