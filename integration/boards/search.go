package boards

import (
	"context"
	"log"
	"os"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/searchshared"
	"github.com/theapemachine/errnie"
)

type SearchWiki struct {
	conn     search.Client
	projects []core.TeamProjectReference
}

func NewSearchWiki(ctx context.Context) (*SearchWiki, error) {
	client, err := search.NewClient(
		ctx, azuredevops.NewPatConnection(
			os.Getenv("AZURE_DEVOPS_ORG_URL"),
			os.Getenv("AZURE_DEVOPS_PAT"),
		),
	)

	if err != nil {
		return nil, errnie.Error(err)
	}

	return &SearchWiki{
		conn:     client,
		projects: []core.TeamProjectReference{},
	}, nil
}

func (searchWiki *SearchWiki) Search(ctx context.Context, query string) (err error) {
	var (
		responseValue *searchshared.WikiSearchResponse
		index         = 0
	)

	if responseValue, err = searchWiki.conn.FetchWikiSearchResults(
		ctx, search.FetchWikiSearchResultsArgs{
			Request: &searchshared.WikiSearchRequest{
				SearchText: &query,
			},
			Project: searchWiki.projects[0].Name,
		},
	); err != nil {
		return errnie.Error(err)
	}

	for _, result := range *responseValue.Results {
		log.Printf("Result[%v] = %v", index, result.Path)
		index++
	}

	return
}
