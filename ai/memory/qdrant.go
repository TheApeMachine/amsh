package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
	"github.com/theapemachine/amsh/errnie"
)

type Qdrant struct {
	client     *qdrant.Client
	collection string
	dimension  uint64
}

func NewQdrant(collection string, dimension uint64) *Qdrant {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		// Handle error (you might want to return an error from this function)
		panic(err)
	}
	return &Qdrant{
		client:     client,
		collection: collection,
		dimension:  dimension,
	}
}

func (q *Qdrant) Read(p []byte) (n int, err error) {
	ctx := context.Background()

	// Create a vector of the correct dimension filled with zeros
	vector := make([]float32, q.dimension)
	var limit uint64 = 1

	request := &qdrant.QueryPoints{
		CollectionName: q.collection,
		Query:          qdrant.NewQuery(vector...),
		Limit:          &limit,
	}

	response, err := q.client.Query(ctx, request)
	if err != nil {
		return 0, fmt.Errorf("failed to query points: %v", err)
	}

	if len(response) > 0 {
		point := response[0]
		data, err := json.Marshal(point)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal point: %v", err)
		}
		n = copy(p, data)
	}

	return n, nil
}

func (q *Qdrant) Write(p []byte) (n int, err error) {
	ctx := context.Background()

	var point qdrant.PointStruct
	if err := json.Unmarshal(p, &point); err != nil {
		return 0, fmt.Errorf("failed to unmarshal point: %v", err)
	}

	request := &qdrant.UpsertPoints{
		CollectionName: q.collection,
		Points:         []*qdrant.PointStruct{&point},
	}

	_, err = q.client.Upsert(ctx, request)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert point: %v", err)
	}

	return len(p), nil
}

func (q *Qdrant) Close() error {
	// Qdrant client doesn't have a Close method, so we'll just return nil
	return nil
}

func (q *Qdrant) EnsureCollection() error {
	ctx := context.Background()

	request := &qdrant.CreateCollection{
		CollectionName: q.collection,
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:     q.dimension,
					Distance: qdrant.Distance_Cosine,
				},
			},
		},
	}

	if err := q.client.CreateCollection(ctx, request); err != nil {
		return errnie.Error(err)
	}

	return nil
}
