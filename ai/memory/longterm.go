package memory

import (
	"errors"

	"github.com/theapemachine/amsh/data"
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

// Read reads data from long-term memory.
func (lt *LongTerm) Read(p []byte) (n int, err error) {
	artifact := data.Unmarshal(p)
	if artifact == nil {
		return 0, errors.New("failed to unmarshal artifact")
	}

	scope := artifact.Peek("scope")
	switch scope {
	case "vector":
		return lt.qdrantClient.Read(p)
	case "graph":
		return lt.neo4jClient.Read(p)
	default:
		return 0, errors.New("invalid long-term memory scope")
	}
}

// Write writes data to long-term memory.
func (lt *LongTerm) Write(p []byte) (n int, err error) {
	artifact := data.Unmarshal(p)
	if artifact == nil {
		return 0, errors.New("failed to unmarshal artifact")
	}

	scope := artifact.Peek("scope")
	switch scope {
	case "vector":
		return lt.qdrantClient.Write(p)
	case "graph":
		return lt.neo4jClient.Write(p)
	default:
		return 0, errors.New("invalid long-term memory scope")
	}
}

// Close closes both the Neo4j and Qdrant clients.
func (lt *LongTerm) Close() error {
	if err := lt.neo4jClient.Close(); err != nil {
		return err
	}
	if err := lt.qdrantClient.Close(); err != nil {
		return err
	}
	return nil
}
