package memory

import (
	"context"
	"log"
	"net/url"

	"github.com/theapemachine/amsh/errnie"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

type Qdrant struct {
	ctx        context.Context
	client     *qdrant.Store
	embedder   embeddings.Embedder
	collection string
	dimension  uint64
}

func NewQdrant(collection string, dimension uint64) *Qdrant {
	ctx := context.Background()

	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	url, err := url.Parse("http://localhost:6333")

	if err != nil {
		url, err = url.Parse("http://qdrant:6333")
	}

	if err != nil {
		errnie.Error(err)
	}

	client, err := qdrant.New(
		qdrant.WithURL(*url),
		qdrant.WithCollectionName(collection),
		qdrant.WithEmbedder(e),
	)
	if err != nil {
		errnie.Error(err)
	}
	return &Qdrant{
		ctx:        ctx,
		client:     &client,
		embedder:   e,
		collection: collection,
		dimension:  dimension,
	}
}

func (q *Qdrant) AddDocuments(docs []schema.Document) error {
	_, err := q.client.AddDocuments(q.ctx, docs)
	return errnie.Error(err)
}

func (q *Qdrant) SimilaritySearch(query string, k int, opts ...vectorstores.Option) ([]schema.Document, error) {
	docs, err := q.client.SimilaritySearch(q.ctx, query, k, opts...)
	return docs, errnie.Error(err)
}

type QdrantResult struct {
	Metadata map[string]any `json:"metadata"`
	Content  string         `json:"content"`
}

func (q *Qdrant) Query(query string) ([]map[string]interface{}, error) {
	docs, err := q.client.SimilaritySearch(q.ctx, query, 1)
	if errnie.Error(err) != nil {
		return nil, errnie.Error(err)
	}

	var results []map[string]interface{}

	for _, doc := range docs {
		results = append(results, map[string]interface{}{
			"metadata": doc.Metadata,
			"content":  doc.PageContent,
		})
	}

	return results, nil
}

func (q *Qdrant) Add(docs []string) ([]string, error) {
	_, err := q.embedder.EmbedDocuments(q.ctx, docs)
	if errnie.Error(err) != nil {
		return nil, errnie.Error(err)
	}

	documents := make([]schema.Document, len(docs))
	for i, doc := range docs {
		documents[i] = schema.Document{
			PageContent: doc,
		}
	}

	return q.client.AddDocuments(q.ctx, documents)
}
