package memory

import (
	"errors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/tmc/langchaingo/schema"
)

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

func (lt *LongTerm) AddDocuments(docs []schema.Document) error {
	return lt.qdrantClient.AddDocuments(docs)
}

func (lt *LongTerm) Write(storeType string, query string) (neo4j.ResultWithContext, error) {
	switch storeType {
	case "graph":
		return lt.neo4jClient.Write(query)
	default:
		return nil, errors.New("invalid long-term memory store type")
	}
}
