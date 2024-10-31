package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v3/client"
	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/errnie"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

type Qdrant struct {
	ctx        context.Context     `json:"-"`
	client     *qdrant.Store       `json:"-"`
	embedder   embeddings.Embedder `json:"-"`
	collection string              `json:"-"`
	dimension  uint64              `json:"-"`
	ToolName   string              `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool that must be 'qdrant',enum=qdrant"`
	Operation  string              `json:"operation" jsonschema:"title=Operation,description=The operation to perform,enum=add,enum=query,required"`
	Question   string              `json:"question" jsonschema:"title=Question,description=The search query for similarity search (required for 'query' operation)"`
	Documents  []string            `json:"documents" jsonschema:"title=Documents,description=The documents to add (required for 'add' operation)"`
}

// Use implements the Tool interface
func (qdrant *Qdrant) Use(ctx context.Context, args map[string]any) string {
	switch qdrant.Operation {
	case "add":
		if docs, ok := args["documents"].([]string); ok {
			ids, err := qdrant.Add(docs)
			if err != nil {
				return err.Error()
			}
			return fmt.Sprintf("Successfully added documents with IDs: %v", ids)
		}
		return "Invalid documents format"

	case "query":
		if query, ok := args["query"].(string); ok {
			results, err := qdrant.Query(query)
			if err != nil {
				return err.Error()
			}
			// Convert results to JSON string
			jsonResults, err := json.Marshal(results)
			if err != nil {
				return err.Error()
			}
			return string(jsonResults)
		}
		return "Invalid query format"

	default:
		return "Unsupported operation"
	}
}

// GenerateSchema implements the Tool interface
func (qdrant *Qdrant) GenerateSchema() string {
	schema := jsonschema.Reflect(&Qdrant{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
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

	url, err := url.Parse(os.Getenv("QDRANT_URL"))

	if err != nil {
		errnie.Error(err)
	}

	createCollectionIfNotExists(collection, url, dimension)

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
	// Perform the similarity search with the options
	docs, err := q.client.SimilaritySearch(q.ctx, query, 1, vectorstores.WithScoreThreshold(0.7))
	if errnie.Error(err) != nil {
		return nil, err
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
		return nil, err
	}

	documents := make([]schema.Document, len(docs))
	for i, doc := range docs {
		documents[i] = schema.Document{
			PageContent: doc,
		}
	}

	return q.client.AddDocuments(q.ctx, documents)
}

/*
createCollectionIfNotExists uses an HTTP PUT call to create a collection if it does not exist.
*/
func createCollectionIfNotExists(collection string, uri *url.URL, dimension uint64) error {
	var (
		response *client.Response
		err      error
	)

	// First we do a GET call to check if the collection exists
	if response, err = client.Get(uri.String() + "/collections/" + collection); errnie.Error(err) != nil {
		return errnie.Error(err)
	}

	if response.StatusCode() == 404 {
		// Prepare the request body for creating a new collection
		requestBody := map[string]interface{}{
			"name": collection,
			"vectors": map[string]interface{}{
				"size":     dimension,
				"distance": "Cosine",
			},
		}

		if response, err = client.Put(uri.String()+"/collections/"+collection, client.Config{
			Header: map[string]string{
				"Content-Type": "application/json",
			},
			Body: requestBody,
		}); errnie.Error(err) != nil {
			return errnie.Error(err)
		}
	}

	errnie.Debug(response.String())

	return nil
}
