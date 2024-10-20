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
	// Initialize the messaging queue
	queue := twoface.NewQueue()

	// Initialize the worker manager
	builder := mastercomputer.NewBuilder()

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

	message := data.New(utils.NewName(), "task", "managing", []byte("Why does Deep Thought claim the answer to life the universe and everything is 42?"))
	message.Poke("chain", "test")

	queue.PubCh <- *message

	log.Println("Waiting for workers to finish...")
	builder.Wait()
	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}
