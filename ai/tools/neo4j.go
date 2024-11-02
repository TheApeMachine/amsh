package tools

import (
	"context"
	"encoding/json"

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
	return string(errnie.SafeMust(func() ([]byte, error) {
		return json.MarshalIndent(schema, "", "  ")
	}))
}

// NewNeo4j creates a new Neo4j client.
func NewNeo4j() *Neo4j {
	ctx := context.Background()

	client := errnie.SafeMust(func() (neo4j.DriverWithContext, error) {
		return neo4j.NewDriverWithContext("neo4j://localhost:7687", neo4j.BasicAuth("neo4j", "securepassword", ""))
	})

	errnie.MustVoid(client.VerifyConnectivity(ctx))
	return &Neo4j{client: client}
}

/*
Query executes a Cypher query on the Neo4j database and returns the results.
*/
func (n *Neo4j) Query(query string) (out []map[string]interface{}, err error) {
	ctx := context.Background()
	session := n.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result := errnie.SafeMust(func() (neo4j.ResultWithContext, error) {
		return session.Run(ctx, query, nil)
	})

	var records []map[string]interface{}
	for result.Next(ctx) {
		record := result.Record()
		records = append(records, record.Values[0].(neo4j.Node).Props)
	}

	errnie.MustVoid(result.Err())
	return records, nil
}

func (n *Neo4j) Write(query string) neo4j.ResultWithContext {
	ctx := context.Background()
	session := n.client.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	return errnie.SafeMust(func() (neo4j.ResultWithContext, error) {
		return session.Run(ctx, query, nil)
	})
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
		records := errnie.SafeMust(func() ([]map[string]interface{}, error) {
			return neo4j.Query(args["cypher"].(string))
		})
		result := errnie.SafeMust(func() ([]byte, error) {
			return json.Marshal(records)
		})
		return string(result)

	case "write":
		return neo4j.Write(args["query"].(string)).Err().Error()

	default:
		return "Unsupported operation"
	}
}
