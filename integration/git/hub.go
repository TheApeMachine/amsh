package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/errnie"
	"golang.org/x/oauth2"
)

type Hub struct {
	ToolName    string `json:"tool" jsonschema:"title=GitHub,description=A tool for interacting with GitHub,enum=github,required"`
	Operation   string `json:"operation" jsonschema:"title=Operation,description=The operation to perform,enum=search_code,required"`
	SearchQuery string `json:"search_query" jsonschema:"title=Search Query,description=The query to search for,required"`
	client      *github.Client
}

func NewHub() *Hub {
	token := os.Getenv("GITHUB_PAT")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &Hub{
		client: client,
	}
}

func (h *Hub) Use(ctx context.Context, args map[string]any) string {
	switch h.Operation {
	case "search_code":
		results, err := h.SearchCode(ctx, h.SearchQuery)
		if err != nil {
			return err.Error()
		}

		out := []string{}

		for _, result := range results {
			// Check the total length of the output
			if len(out)+len(result.Repository)+len(result.Path)+len(result.Content) > 1024 {
				break
			}

			out = append(out, fmt.Sprintf("Repository: %s\nPath: %s\nContent: %s\n", result.Repository, result.Path, result.Content))
		}

		return strings.Join(out, "\n")
	}

	return ""
}

func (h *Hub) GenerateSchema() string {
	schema := jsonschema.Reflect(&Hub{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

type CodeSearchResult struct {
	Repository string `json:"repository"`
	Path       string `json:"path"`
	Content    string `json:"content"`
}

func (h *Hub) SearchCode(ctx context.Context, query string) ([]CodeSearchResult, error) {
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 3},
	}

	result, _, err := h.client.Search.Code(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	var searchResults []CodeSearchResult
	for _, item := range result.CodeResults {
		content, err := h.getFileContent(ctx, *item.Repository.Owner.Login, *item.Repository.Name, *item.Path)
		if err != nil {
			// Log the error but continue with other results
			fmt.Printf("Error getting content for %s/%s/%s: %v\n", *item.Repository.Owner.Login, *item.Repository.Name, *item.Path, err)
			continue
		}

		content = strings.TrimSpace(content)

		if len(content) > 2000 {
			content = content[:2000]
		}

		searchResults = append(searchResults, CodeSearchResult{
			Repository: *item.Repository.FullName,
			Path:       *item.Path,
			Content:    content,
		})
	}

	return searchResults, nil
}

func (h *Hub) getFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	fileContent, _, _, err := h.client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		return "", err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", err
	}

	return content, nil
}
