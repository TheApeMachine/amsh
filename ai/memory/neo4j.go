package memory

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/errnie"
)

// Neo4j represents the Neo4j client.
type Neo4j struct {
	client neo4j.DriverWithContext
}

// NewNeo4j creates a new Neo4j client.
func NewNeo4j() *Neo4j {
	ctx := context.Background()
	// uri := viper.GetString("neo4j.uri")
	// username := viper.GetString("neo4j.username")
	// password := viper.GetString("neo4j.password")
	var (
		client neo4j.DriverWithContext
		err    error
	)

	client, err = neo4j.NewDriverWithContext("neo4j://localhost:7687", neo4j.BasicAuth("neo4j", "securepassword", ""))

	if err != nil {
		errnie.Error(err)
		client, err = neo4j.NewDriverWithContext("neo4j://neo4j:7687", neo4j.BasicAuth("neo4j", "securepassword", ""))
	}

	// Verify connectivity
	if err := client.VerifyConnectivity(ctx); err != nil {
		errnie.Error(err)
	}

	return &Neo4j{client: client}
}

/*
Query executes a Cypher query on the Neo4j database and returns the results.
*/
func (n *Neo4j) Query(query string) ([]map[string]interface{}, error) {
	ctx := context.Background()
	session := n.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	var records []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()
		records = append(records, record.Values[0].(neo4j.Node).Props)
	}

	if err = result.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// Close closes the Neo4j client connection.
func (n *Neo4j) Close() error {
	ctx := context.Background()
	return n.client.Close(ctx)
}

func (n *Neo4j) Call(args map[string]any) (string, error) {
	return "", nil
}

func (n *Neo4j) Schema() openai.ChatCompletionToolParam {
	return MakeTool(
		"neo4j",
		"Manage your long-term graph memory by storing and querying nodes and relationships.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]string{
					"type":        "string",
					"description": "The Cypher query to execute on the Neo4j database",
				},
			},
			"required": []string{"query"},
		},
	)
}
