package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/spf13/cobra"
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
		)

		defer pipeline.Save()

		// Open a new log file for writing.
		logFileName := fmt.Sprintf("logs/run_%s.md", time.Now().Format("2006-01-02_15-04-05"))
		logFile, err := os.Create(logFileName)
		if err != nil {
			return err
		}
		defer logFile.Close()

		for chunk := range pipeline.Generate() {
			fmt.Print(chunk)
			logFile.WriteString(stripansi.Strip(chunk))
		}

		return nil
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
