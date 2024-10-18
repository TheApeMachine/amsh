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
	manager := mastercomputer.NewManager()
	builder := mastercomputer.NewBuilder(ctx, manager)

	for _, agent := range []mastercomputer.WorkerType{
		mastercomputer.WorkerTypeManager,
		mastercomputer.WorkerTypeReasoner,
		mastercomputer.WorkerTypeVerifier,
		mastercomputer.WorkerTypeCommunicator,
		mastercomputer.WorkerTypeResearcher,
		mastercomputer.WorkerTypeExecutor,
	} {
		worker := builder.NewWorker(agent)
		worker.Start()
	}

	// Simulate an external prompt being broadcasted
	externalPrompt := data.New(utils.NewName(), "task", "managing", []byte{})
	externalPrompt.Poke("id", utils.NewID())
	externalPrompt.Poke("payload", "Solve the riddle: In a fruit's sweet name, I'm hidden three, A triple threat within its juicy spree. Find me and you'll discover a secret delight.")

	queue.Publish(externalPrompt)

	log.Println("Waiting for workers to finish...")
	manager.Wait()
	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}
