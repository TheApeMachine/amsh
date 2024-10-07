package memory

import (
	"log"

	"github.com/qdrant/go-client/qdrant"
)

type Qdrant struct {
	ID     string
	client *qdrant.Client
}

func NewQdrant(ID string) *Qdrant {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6333,
	})

	if err != nil {
		log.Fatalf("Failed to create Qdrant client: %v", err)
		return nil
	}

	return &Qdrant{ID: ID, client: client}
}

func (qdrant *Qdrant) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (qdrant *Qdrant) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (qdrant *Qdrant) Close() error {
	return nil
}
