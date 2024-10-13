package memory

import (
	"context"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

type Qdrant struct {
	ctx        context.Context
	client     *qdrant.Store
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

	url, err := url.Parse("http://localhost:6334")
	if err != nil {
		log.Fatal(err)
	}

	client, err := qdrant.New(
		qdrant.WithURL(*url),
		qdrant.WithCollectionName(collection),
		qdrant.WithEmbedder(e),
	)
	if err != nil {
		// Handle error (you might want to return an error from this function)
		panic(err)
	}
	return &Qdrant{
		ctx:        ctx,
		client:     &client,
		collection: collection,
		dimension:  dimension,
	}
}

func (q *Qdrant) Read(p []byte) (n int, err error) {
	artifact := data.Empty
	artifact = artifact.Unmarshal(p)

	var docs []schema.Document
	if docs, err = q.client.SimilaritySearch(q.ctx, artifact.Peek("query"), 1); errnie.Error(err) != nil {
		return 0, errnie.Error(err)
	}

	builder := strings.Builder{}

	for _, doc := range docs {
		builder.WriteString(doc.PageContent)
	}

	artifact.Poke("payload", builder.String())
	buf := artifact.Marshal()

	copy(p, buf)

	return len(p), io.EOF
}

func (q *Qdrant) Write(p []byte) (n int, err error) {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Qdrant vector store.
	url, err := url.Parse("http://localhost:6334")
	if err != nil {
		log.Fatal(err)
	}
	store, err := qdrant.New(
		qdrant.WithURL(*url),
		qdrant.WithCollectionName(q.collection),
		qdrant.WithEmbedder(e),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Add documents to the Qdrant vector store.
	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{
			PageContent: "A city in texas",
			Metadata: map[string]any{
				"area": 3251,
			},
		},
		{
			PageContent: "A country in Asia",
			Metadata: map[string]any{
				"area": 2342,
			},
		},
		{
			PageContent: "A country in South America",
			Metadata: map[string]any{
				"area": 432,
			},
		},
		{
			PageContent: "An island nation in the Pacific Ocean",
			Metadata: map[string]any{
				"area": 6531,
			},
		},
		{
			PageContent: "A mountainous country in Europe",
			Metadata: map[string]any{
				"area": 1211,
			},
		},
		{
			PageContent: "A lost city in the Amazon",
			Metadata: map[string]any{
				"area": 1223,
			},
		},
		{
			PageContent: "A city in England",
			Metadata: map[string]any{
				"area": 4324,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	return len(p), nil
}

func (q *Qdrant) Close() error {
	// Qdrant client doesn't have a Close method, so we'll just return nil
	return nil
}
