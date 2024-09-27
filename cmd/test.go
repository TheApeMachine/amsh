package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/ui"
)

/*
testCmd acts as an entry point for testing new features.
*/
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the service with the ~/.amsh/config.yml config values.",
	Long:  testtxt,
	RunE: func(_ *cobra.Command, _ []string) (err error) {
		fmt.Print(ui.Logo)

		pipeline := ai.NewPipeline(
			context.Background(),
			ai.NewConn(),
			viper.GetViper().GetStringSlice("ai.prompt.system.steps")...,
		)

		// Open a new log file for writing.
		var logFile *os.File
		if logFile, err = os.Create("logs/run" + time.Now().Format("2006-01-02 15:04:05") + ".md"); err != nil {
			return err
		}

		defer logFile.Close()

		for chunk := range pipeline.Generate() {
			fmt.Print(chunk)
			logFile.WriteString(chunk)
		}

		return
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	errnie.Debug("Test command initialized")
}

/*
testtxt lives here to keep the command definition section cleaner.
*/
var testtxt = `
test new features
`
