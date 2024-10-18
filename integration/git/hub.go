package git

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

type Hub struct {
	client *github.Client
}

func NewHub() *Hub {
	token := os.Getenv("GITHUB_TOKEN")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &Hub{
		client: client,
	}
}

type CodeSearchResult struct {
	Repository string `json:"repository"`
	Path       string `json:"path"`
	Content    string `json:"content"`
}

func (h *Hub) SearchCode(ctx context.Context, query string) ([]CodeSearchResult, error) {
	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 30}, // Limit to 30 results
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
