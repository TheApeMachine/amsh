package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/internal/app"
)

var (
	path string
)

var rootCmd = &cobra.Command{
	Use:   "amsh",
	Short: "A minimal shell and vim-like text editor with A.I. capabilities",
	Long:  roottxt,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(
			app.NewModel(path),
			tea.WithAltScreen(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Println("Error while running program:", err)
			os.Exit(1)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to open")
}

const roottxt = `amsh v0.0.1
A minimal shell and vim-like text editor written in Go, with integrated A.I. capabilities.
Different from other A.I. integrations, it uses multiple A.I. models that engage independently
in conversation with each other and the user, improving the developer experience and providing
a more human-like interaction.
`
