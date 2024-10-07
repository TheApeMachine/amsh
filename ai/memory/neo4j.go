package memory

import (
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/spf13/viper"
)

type Neo4j struct {
	ID     string
	client neo4j.DriverWithContext
}

func NewNeo4j(ID string) *Neo4j {
	client, err := neo4j.NewDriverWithContext(
		"neo4j://neo4j:7687",
		neo4j.BasicAuth(
			viper.GetViper().GetString("neo4j.username"),
			viper.GetViper().GetString("neo4j.password"),
			"",
		),
	)

	if err != nil {
		log.Fatalf("Failed to create Neo4j client: %v", err)
		return nil
	}

	return &Neo4j{ID: ID, client: client}
}

func (neo4j *Neo4j) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (neo4j *Neo4j) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (neo4j *Neo4j) Close() error {
	return nil
}
