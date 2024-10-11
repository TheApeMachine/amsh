package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/spf13/viper"
)

type Neo4j struct {
	client neo4j.DriverWithContext
}

func NewNeo4j() *Neo4j {
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

	return &Neo4j{client: client}
}

func (conn *Neo4j) Read(p []byte) (n int, err error) {
	ctx := context.Background()
	session := conn.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, "MATCH (n) RETURN n LIMIT 1", nil)
	if err != nil {
		return 0, err
	}

	if result.Next(ctx) {
		record := result.Record()
		node, ok := record.Values[0].(neo4j.Node)
		if !ok {
			return 0, fmt.Errorf("expected neo4j.Node, got %T", record.Values[0])
		}
		data, err := json.Marshal(node.Props)
		if err != nil {
			return 0, err
		}
		n = copy(p, data)
	}

	return n, result.Err()
}

func (conn *Neo4j) Write(p []byte) (n int, err error) {
	ctx := context.Background()
	session := conn.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	var data map[string]interface{}
	if err := json.Unmarshal(p, &data); err != nil {
		return 0, err
	}

	result, err := session.Run(ctx, "CREATE (n:Node $props) RETURN n", map[string]interface{}{"props": data})
	if err != nil {
		return 0, err
	}

	if result.Next(ctx) {
		record := result.Record()
		if _, ok := record.Values[0].(neo4j.Node); !ok {
			return 0, fmt.Errorf("expected neo4j.Node, got %T", record.Values[0])
		}
		n = len(p)
	}

	return n, result.Err()
}

func (conn *Neo4j) Close() error {
	ctx := context.Background()
	return conn.client.Close(ctx)
}
