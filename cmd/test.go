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

	// Simulate an external prompt being broadcasted
	externalPrompt := data.New(utils.NewName(), "manager", "managing", []byte{})
	externalPrompt.Poke("stage", "ingress")
	externalPrompt.Poke("id", utils.NewID())
	externalPrompt.Poke("payload", "We need to design a new testing methodology to measure employee wellbeing in the workplace. It has to both highly robust, but also engaging, becausee surveys are not proving to be very effective.")

	queue.Publish(externalPrompt)

	log.Println("Waiting for workers to finish...")
	builder.Wait()
	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}
