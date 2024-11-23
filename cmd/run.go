package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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

		// Then start the program
		if _, err := tea.NewProgram(tui.NewApp(), tea.WithAltScreen()).Run(); err != nil {
			fmt.Println("Error while running program:", err)
			os.Exit(1)
		}

		return nil
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
