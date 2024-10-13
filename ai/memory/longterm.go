package memory

import (
	"io"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
)

type LongTerm struct {
	neo4jClient  *Neo4j
	qdrantClient *Qdrant
}

func NewLongTerm(agentID string) *LongTerm {
	return &LongTerm{
		neo4jClient:  NewNeo4j(),
		qdrantClient: NewQdrant(agentID, 3),
	}
}

func (store *LongTerm) Read(p []byte) (n int, err error) {
	artifact := data.Empty
	artifact.Write(p)

	switch artifact.Peek("scope") {
	case "vector":
		return store.qdrantClient.Read(p)
	case "graph":
		return store.neo4jClient.Read(p)
	default:
		errnie.Warn("invalid memory store called")
	}

	return len(p), io.EOF
}

func (store *LongTerm) Write(p []byte) (n int, err error) {
	artifact := data.Empty
	artifact = artifact.Unmarshal(p)

	switch artifact.Peek("scope") {
	case "vector":
		io.Copy(store.qdrantClient, artifact)
	case "graph":
		io.Copy(store.neo4jClient, artifact)
	}

	return len(p), nil
}
