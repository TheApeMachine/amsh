package memory

import (
	"context"
	"errors"
	"log"

	"github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
)

type SemanticMemory struct {
	client qdrant.PointsClient
}

func NewSemanticMemory() *SemanticMemory {
	conn, err := grpc.Dial("qdrant:6334", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Qdrant: %v", err)
	}

	client := qdrant.NewPointsClient(conn)
	return &SemanticMemory{client: client}
}

func (sm *SemanticMemory) CreateCollection(ctx context.Context, name string, vectorSize int) error {
	return errors.New("not implemented")
}

func (sm *SemanticMemory) StoreEmbedding(ctx context.Context, collectionName string, vector []float32, payload map[string]interface{}) error {
	return errors.New("not implemented")
}

func (sm *SemanticMemory) QuerySimilar(ctx context.Context, collectionName string, vector []float32, topK int) ([]*qdrant.ScoredPoint, error) {
	return nil, errors.New("not implemented")
}
