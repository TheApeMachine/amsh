package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

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

		v := viper.GetViper()

		prompt.Poke("system", v.GetString("ai.prompt.system"))
		prompt.Poke("guidelines", v.GetString("ai.prompt.guidelines"))
		prompt.Poke("toolset", v.GetString("system"))
		prompt.Poke("user", "Have a look at the following code, and make it better.\n\n"+buf.String())

		systems := []*mastercomputer.Worker{
			mastercomputer.NewWorker(cmd.Context(), prompt).Initialize(cmd.Context()),
		}

		for _, system := range systems {
			if !system.OK {
				errnie.Warn("skipping system")
				continue
			}
			system.Process(context.Background())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
