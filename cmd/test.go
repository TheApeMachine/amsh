package cmd

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/mastercomputer"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		errnie.Trace()

		system := viper.GetViper().GetString("ai.prompt.system")
		system = strings.ReplaceAll(system, "{id}", "test")
		system = strings.ReplaceAll(system, "{guidelines}", viper.GetViper().GetString("ai.prompt.guidelines"))
		user := viper.GetViper().GetString("ai.prompt.reasoning")

		workers := make([]*mastercomputer.Worker, 2)

		for i := range workers {
			workers[i] = mastercomputer.NewWorker(
				cmd.Context(),
				data.New(
					"test", "buffer", "setup", []byte{},
				).Poke(
					"system", system,
				).Poke(
					"user", user,
				).Poke(
					"format", "reasoning",
				).Poke(
					"toolset", "reasoning",
				),
			).Initialize()
		}

		go func() {
			time.Sleep(time.Second * 5)
			workers[0].Test(data.New(
				"test", "test", "broadcast", []byte("Hello, world!"),
			))
		}()

		for _, worker := range workers {
			if err = worker.Process(); err != nil {
				return err
			}

			if !worker.OK {
				return errors.New("worker not ok")
			}
		}

		select {}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
