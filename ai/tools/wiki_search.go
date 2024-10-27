package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/searchshared"
	"github.com/theapemachine/amsh/ai/types"
)

type WikiSearchTool struct {
	client      search.Client
	projectName string
}

func NewWikiSearchTool(ctx context.Context) (*WikiSearchTool, error) {
	organizationURL := os.Getenv("AZDO_ORG_URL")
	personalAccessToken := os.Getenv("AZDO_PAT")
	projectName := os.Getenv("AZDO_PROJECT_NAME")

	if organizationURL == "" || personalAccessToken == "" || projectName == "" {
		return nil, errors.New("azure devops configuration is missing")
	}

	connection := azuredevops.NewPatConnection(organizationURL, personalAccessToken)
	client, err := search.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure DevOps search client: %w", err)
	}

	return &WikiSearchTool{
		client:      client,
		projectName: projectName,
	}, nil
}

func (w *WikiSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	keywords, err := getStringArg(args, "keywords", "")
	if err != nil || keywords == "" {
		return "", errors.New("keywords are required for search operation")
	}

	request := searchshared.WikiSearchRequest{
		SearchText: &keywords,
	}

	response, err := w.client.FetchWikiSearchResults(ctx, search.FetchWikiSearchResultsArgs{
		Request: &request,
	})
	if err != nil {
		return "", fmt.Errorf("failed to search wiki: %w", err)
	}

	if len(*response.Results) == 0 {
		return "No wiki pages found for the given keywords.", nil
	}

	var output strings.Builder
	output.WriteString("Wiki Search Results:\n")
	for _, result := range *response.Results {
		output.WriteString(
			fmt.Sprintf("- Title: %s\n  Path: %s\n  URL: %s\n\n", *result.Wiki.Name, *result.Path, *result.Wiki.MappedPath))
	}

	return output.String(), nil
}

func (w *WikiSearchTool) GetSchema() types.ToolSchema {
	return types.ToolSchema{
		Name:        "wiki_search",
		Description: "Search Azure DevOps Wiki pages.",
		Parameters: map[string]interface{}{
			"keywords": map[string]interface{}{
				"type":        "string",
				"description": "The keywords to search for in the wiki.",
			},
		},
	}
}
