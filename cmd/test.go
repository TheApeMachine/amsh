package cmd

import (
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/mastercomputer"
	"github.com/theapemachine/amsh/utils"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		errnie.Trace()
		v := viper.GetViper()

		workers := []*mastercomputer.Worker{}
		wg := sync.WaitGroup{}

		for i := 0; i < 2; i++ {
			ID := utils.NewID()
			role := "reasoner"
			workload := "reasoning"
			workers = append(workers, mastercomputer.NewWorker(
				cmd.Context(),
				&wg,
				data.New(
					utils.NewID(), "buffer", "setup", []byte{},
				).Poke(
					"system", utils.ReplaceWith(v.GetString("ai.prompt.system"), [][]string{
						{"id", ID},
						{"role", role},
						{"guidelines", v.GetString("ai.prompt.guidelines")},
					}),
				).Poke(
					"user", utils.ReplaceWith("Welcome aboard, and good luck, Agent {id}!", [][]string{
						{"id", ID},
					}),
				).Poke(
					"workload", workload,
				),
			))
		}

		for _, worker := range workers {
			wg.Add(1)
			worker.Initialize()
		}

		wg.Wait()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
