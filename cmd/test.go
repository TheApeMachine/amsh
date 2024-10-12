package cmd

import (
	"errors"
	"io"
	"time"

	"github.com/spf13/cobra"
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

		worker := mastercomputer.NewWorker(cmd.Context(), data.New("test", "test", "test", []byte{}))
		worker.Initialize()

		if _, err = io.Copy(worker, data.New(
			"test", "initialize", "reasoner", nil,
		).Poke(
			"format", "reasoning",
		).Poke(
			"toolset", "reasoning",
		)); err != nil {
			return err
		}

		go func() {
			time.Sleep(time.Second * 2)
			worker.Test(data.New(
				"test", "test", "broadcast", []byte("Hello, world!"),
			))
		}()

		if err = worker.Process(); err != nil {
			return err
		}

		if !worker.OK {
			return errors.New("worker not ok")
		}

		return errors.New(worker.Error())
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
