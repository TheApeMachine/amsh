package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/searchshared"
	"github.com/theapemachine/amsh/errnie"
)

type Wiki struct {
	client      search.Client
	projectName string
}

func NewWiki() *Wiki {
	organizationURL := os.Getenv("AZDO_ORG_URL")
	personalAccessToken := os.Getenv("AZDO_PAT")
	projectName := os.Getenv("AZDO_PROJECT_NAME")

	if organizationURL == "" || personalAccessToken == "" || projectName == "" {
		return nil
	}

	connection := azuredevops.NewPatConnection(organizationURL, personalAccessToken)
	client, err := search.NewClient(context.Background(), connection)
	if err != nil {
		return nil
	}

	return &Wiki{
		client:      client,
		projectName: projectName,
	}
}

func (wiki *Wiki) GenerateSchema() string {
	schema := jsonschema.Reflect(&Wiki{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (w *Wiki) Use(ctx context.Context, args map[string]any) string {
	keywords, err := getStringArg(args, "keywords", "")
	if err != nil || keywords == "" {
		return errnie.Error(errors.New("keywords are required for search operation")).Error()
	}

	request := searchshared.WikiSearchRequest{
		SearchText: &keywords,
	}

	response, err := w.client.FetchWikiSearchResults(ctx, search.FetchWikiSearchResultsArgs{
		Request: &request,
	})
	if err != nil {
		return errnie.Error(fmt.Errorf("failed to search wiki: %w", err)).Error()
	}

	if len(*response.Results) == 0 {
		return "No wiki pages found for the given keywords."
	}

	var output strings.Builder
	output.WriteString("Wiki Search Results:\n")
	for _, result := range *response.Results {
		output.WriteString(
			fmt.Sprintf("- Title: %s\n  Path: %s\n  URL: %s\n\n", *result.Wiki.Name, *result.Path, *result.Wiki.MappedPath))
	}

	return output.String()
}
