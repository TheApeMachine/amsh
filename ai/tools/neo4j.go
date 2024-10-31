package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/theapemachine/amsh/errnie"
)

// Neo4j represents the Neo4j client.
type Neo4j struct {
	client    neo4j.DriverWithContext `json:"-"`
	ToolName  string                  `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool that must be 'neo4j',enum=neo4j"`
	Operation string                  `json:"operation" jsonschema:"title=Operation,description=The operation to perform,enum=query,enum=write,required"`
	Cypher    string                  `json:"cypher" jsonschema:"title=Cypher,description=The Cypher query to execute,required"`
}

// GenerateSchema implements the Tool interface
func (neo4j *Neo4j) GenerateSchema() string {
	schema := jsonschema.Reflect(&Neo4j{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

// NewNeo4j creates a new Neo4j client.
func NewNeo4j() *Neo4j {
	ctx := context.Background()
	var (
		client neo4j.DriverWithContext
		err    error
	)

	client, err = neo4j.NewDriverWithContext("neo4j://localhost:7687", neo4j.BasicAuth("neo4j", "securepassword", ""))

	if err != nil {
		errnie.Error(err)
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
func (n *Neo4j) Query(query string) (out []map[string]interface{}, err error) {
	ctx := context.Background()
	session := n.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var result neo4j.ResultWithContext

	if result, err = session.Run(ctx, query, nil); err != nil {
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

func (n *Neo4j) Write(query string) (neo4j.ResultWithContext, error) {
	ctx := context.Background()
	session := n.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return session.Run(ctx, query, nil)
}

// Close closes the Neo4j client connection.
func (n *Neo4j) Close() error {
	ctx := context.Background()
	return n.client.Close(ctx)
}

// Use implements the Tool interface
func (neo4j *Neo4j) Use(ctx context.Context, args map[string]any) string {
	switch neo4j.Operation {
	case "query":
		records, err := neo4j.Query(args["cypher"].(string))
		if err != nil {
			return err.Error()
		}
		result, err := json.Marshal(records)
		if err != nil {
			return err.Error()
		}
		return string(result)

	case "write":
		result, err := neo4j.Write(args["query"].(string))
		if err != nil {
			return err.Error()
		}
		// Convert result to string representation
		summary, err := result.Consume(ctx)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("Affected nodes: %d", summary.Counters().NodesCreated())

	default:
		return "Unsupported operation"
	}
}
