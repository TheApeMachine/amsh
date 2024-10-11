package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/mastercomputer"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := data.New(
			"user",
			"prompt",
			"task",
			[]byte{},
		)

		// Get working directory
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Load a file from the file system
		file, err := os.Open(filepath.Join(wd, "tui", "core", "buffer.go"))
		if err != nil {
			return err
		}
		defer file.Close()

		// Read the file to a string
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(file)
		if err != nil {
			return err
		}

		prompt.Poke("system", "You are part of an advanced AI system, called The Ape Machine. When you are presented with code, analyze it deeply then provide an improved version of it.")
		prompt.Poke("guidelines", "If at some point you do not know what to do, you should just say so and ask for help. Never make up an answer, or try to guess at an answer.")
		prompt.Poke("toolset", "system")
		prompt.Poke("user", buf.String())

		systems := []*mastercomputer.Worker{
			mastercomputer.NewWorker(cmd.Context(), prompt),
		}

		for _, system := range systems {
			system.Process(context.Background())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
