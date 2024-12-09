package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui"
)

/*
runCmd acts as an entry point for running the service.
*/
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the service with the ~/.amsh/config.yml config values.",
	Long:  runtxt,
	RunE: func(cmd *cobra.Command, _ []string) (err error) {
		// Set environment variables before anything else
		os.Setenv("NOCONSOLE", "true")
		os.Setenv("LOGFILE", "true")

		// Initialize logger explicitly
		errnie.InitLogger()

		// Create and run the editor
		app := tui.NewApp()
		return app.Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

/*
runtxt provides a long description for the run command.
*/
var runtxt = `
Run the AMSH service using the configuration specified in ~/.amsh/config.yml.
`
