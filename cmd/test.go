package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/mastercomputer"
	"github.com/theapemachine/amsh/twoface"
	"github.com/theapemachine/amsh/utils"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE:  runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Initialize the messaging queue
	queue := twoface.NewQueue()

	// Initialize the worker manager
	manager := mastercomputer.NewWorkerManager()
	builder := mastercomputer.NewBuilder(ctx, manager)

	reasoner := builder.NewWorker("reasoner")
	reasoner.Start()

	// Simulate an external prompt being broadcasted
	externalPrompt := data.New(utils.NewName(), "message", "broadcast", []byte{})
	externalPrompt.Poke("id", utils.NewID())
	externalPrompt.Poke("user", "How many times do we find the letter r in the word strawberry?")

	queue.Publish(externalPrompt)

	log.Println("Waiting for workers to finish...")
	manager.Wait()
	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}
