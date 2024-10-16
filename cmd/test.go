package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	v := viper.GetViper()
	ctx := cmd.Context()

	// Initialize the messaging queue
	queue := twoface.NewQueue()

	// Initialize the worker manager
	manager := mastercomputer.NewWorkerManager()

	// Create workers
	for i := 0; i < 2; i++ {
		ID := utils.NewID()
		role := "reasoner"
		workload := "reasoning"

		artifact := data.New(utils.NewID(), "buffer", "setup", nil)
		artifact.Poke("id", ID)
		artifact.Poke("system", utils.ReplaceWith(v.GetString("ai.prompt.system"), [][]string{
			{"id", ID},
			{"role", role},
			{"guidelines", v.GetString("ai.prompt.guidelines")},
		}))
		artifact.Poke("user", utils.ReplaceWith("Welcome aboard, and good luck, Agent {id}!", [][]string{
			{"id", ID},
		}))
		artifact.Poke("workload", workload)
		artifact.Poke("payload", fmt.Sprintf("Agent %s is ready to go!", ID))

		worker := mastercomputer.NewWorker(ctx, artifact, manager)
		go func() {
			if err := worker.Initialize(); err != nil {
				log.Printf("Worker %s initialization failed: %v", ID, err)
			}
		}()
	}

	// Simulate an external prompt being broadcasted
	externalPrompt := data.New(utils.NewID(), "broadcast", "message", []byte("Solve this complex problem"))
	externalPrompt.Poke("origin", "external")
	externalPrompt.Poke("scope", "broadcast")
	queue.Publish(externalPrompt)

	log.Println("Waiting for workers to finish...")
	manager.Wait()
	return nil
}

func init() {
	rootCmd.AddCommand(testCmd)
}
