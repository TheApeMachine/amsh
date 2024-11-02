package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/theapemachine/amsh/parser"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the Neo4J database",
	Long:  `Seed the Neo4J database with the provided data.`,
	RunE:  runSeed,
}

func runSeed(cmd *cobra.Command, args []string) error {
	parser := parser.NewTreeSitterParser()
	parser.WalkDir("/home/theapemachine/go/src/github.com/fanfactory/gateway")
	return nil
}

func init() {
	rootCmd.AddCommand(seedCmd)
	os.Setenv("NEO4J_URL", "neo4j://localhost:7474")
}
