package memory

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
)

// Neo4j represents the Neo4j client.
type Neo4j struct {
	client neo4j.DriverWithContext
}

// NewNeo4j creates a new Neo4j client.
func NewNeo4j() *Neo4j {
	ctx := context.Background()
	uri := viper.GetString("neo4j.uri")
	username := viper.GetString("neo4j.username")
	password := viper.GetString("neo4j.password")

	client, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		panic("Failed to create Neo4j client: " + err.Error())
	}

	// Verify connectivity
	if err := client.VerifyConnectivity(ctx); err != nil {
		panic("Failed to connect to Neo4j: " + err.Error())
	}

	return &Neo4j{client: client}
}

// Read executes a Cypher query and returns the result.
func (n *Neo4j) Read(p []byte) (int, error) {
	ctx := context.Background()
	session := n.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	artifact := data.Unmarshal(p)
	if artifact == nil {
		return 0, errors.New("failed to unmarshal artifact")
	}

	query := artifact.Peek("cypher")
	if query == "" {
		return 0, errors.New("cypher query is empty")
	}

	result, err := session.Run(ctx, query, nil)
	if err != nil {
		return 0, err
	}

	var records []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()
		records = append(records, record.Values[0].(neo4j.Node).Props)
	}

	if err = result.Err(); err != nil {
		return 0, err
	}

	dataBytes, err := json.Marshal(records)
	if err != nil {
		return 0, err
	}

	copy(p, dataBytes)
	return len(dataBytes), io.EOF
}

// Write executes a Cypher query to write data to the graph.
func (n *Neo4j) Write(p []byte) (int, error) {
	ctx := context.Background()
	session := n.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	artifact := data.Unmarshal(p)
	if artifact == nil {
		return 0, errors.New("failed to unmarshal artifact")
	}

	query := artifact.Peek("cypher")
	if query == "" {
		return 0, errors.New("cypher query is empty")
	}

	_, err := session.Run(ctx, query, nil)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close closes the Neo4j client connection.
func (n *Neo4j) Close() error {
	ctx := context.Background()
	return n.client.Close(ctx)
}
