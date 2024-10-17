package memory

import (
	"errors"

	"github.com/openai/openai-go"
)

// MakeTool reduces boilerplate for creating a tool.
func MakeTool(name, description string, schema openai.FunctionParameters) openai.ChatCompletionToolParam {
	return openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(name),
			Description: openai.String(description),
			Parameters:  openai.F(schema),
		}),
	}
}

// LongTerm represents the long-term memory of a worker.
type LongTerm struct {
	neo4jClient  *Neo4j
	qdrantClient *Qdrant
}

// NewLongTerm creates a new LongTerm memory instance.
func NewLongTerm(agentID string) *LongTerm {
	return &LongTerm{
		neo4jClient:  NewNeo4j(),
		qdrantClient: NewQdrant(agentID, 1536),
	}
}

func (lt *LongTerm) Query(storeType string, query string) ([]map[string]interface{}, error) {
	switch storeType {
	case "graph":
		return lt.neo4jClient.Query(query)
	case "vector":
		return lt.qdrantClient.Query(query)
	default:
		return nil, errors.New("invalid long-term memory store type")
	}
}
