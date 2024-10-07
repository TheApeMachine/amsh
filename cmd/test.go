package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/integration/boards"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run the AI pipeline interactively",
	Long:  `Run the AI pipeline interactively, allowing you to input prompts and see the reasoning process.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := boards.NewService()
		srv.SearchWorkitems(context.Background(), "test")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
