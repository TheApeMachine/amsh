package boards

import (
	"context"
	"log"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/searchshared"
	"github.com/theapemachine/amsh/errnie"
)

func (srv *Service) SearchWiki(ctx context.Context, query string) (err error) {
	var (
		client        search.Client
		projects      []core.TeamProjectReference
		responseValue *searchshared.WikiSearchResponse
		index         = 0
	)

	if client, err = search.NewClient(ctx, azuredevops.NewPatConnection(srv.orgURL, srv.pat)); err != nil {
		return errnie.Error(err)
	}

	if projects, err = srv.GetProjects(ctx); err != nil {
		return errnie.Error(err)
	}

	if responseValue, err = client.FetchWikiSearchResults(ctx, search.FetchWikiSearchResultsArgs{
		Request: &searchshared.WikiSearchRequest{
			SearchText: &query,
		},
		Project: projects[0].Name,
	}); err != nil {
		return errnie.Error(err)
	}

	for _, result := range *responseValue.Results {
		log.Printf("Result[%v] = %v", index, result.Path)
		index++
	}

	return
}
