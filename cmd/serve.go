package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/service"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the amsh service",
	Long:  serveTxt,
	RunE: func(cmd *cobra.Command, _ []string) error {
		os.Setenv("QDRANT_URL", "http://qdrant:6333")
		os.Setenv("NEO4J_URL", "http://neo4j:7474")

		return service.NewHTTPS().Up()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

const serveTxt = `
Starts a fiber 3 HTTP service, using pre-forking to handle requests,
benefiting from additional performance and reduced memory usage due to
Go not having to manage memory across multiple processes.
`
