package memory

import "io"

type Store interface {
	io.ReadWriteCloser
}

type Memory struct {
	ID     string
	stores map[string]Store
}

func NewMemory(ID string) *Memory {
	return &Memory{
		ID: ID,
		stores: map[string]Store{
			"neo4j":  NewNeo4j(ID),
			"qdrant": NewQdrant(ID),
		},
	}
}

func (mem *Memory) Question(question string) (string, error) {
	return "", nil
}

func (mem *Memory) Cypher(cypher string) (string, error) {
	mem.stores["neo4j"].Write([]byte(cypher))
	return "", nil
}
