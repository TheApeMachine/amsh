package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/ui"
)

var description = `
amsh is a custom shell for me. It is not for you.
`

var rootCmd = &cobra.Command{
	Use:   "amsh",
	Short: "My shell/my editor.",
	Long:  description,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := tea.NewProgram(ui.NewState(), tea.WithAltScreen()).Run(); err != nil {
			log.Error(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
