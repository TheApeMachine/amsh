package boards

import (
	"context"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/theapemachine/amsh/errnie"
)

func (srv *Service) SearchWorkitems(ctx context.Context, query string) (err error) {
	var (
		client        search.Client
		responseValue *search.WorkItemSearchResponse
		index         = 0
		skip          = 0
		top           = 10
		filters       = map[string][]string{
			"System.State": {"To Do"},
		}
	)

	if client, err = search.NewClient(ctx, azuredevops.NewPatConnection(srv.orgURL, srv.pat)); err != nil {
		return errnie.Error(err)
	}

	if responseValue, err = client.FetchWorkItemSearchResults(ctx, search.FetchWorkItemSearchResultsArgs{
		Request: &search.WorkItemSearchRequest{
			Filters:    &filters,
			SearchText: &query,
			Skip:       &skip,
			Top:        &top,
		},
		Project: &srv.projectName,
	}); err != nil {
		return errnie.Error(err)
	}

	if responseValue.Results == nil {
		log.Println("No results found")
		return
	}

	for _, result := range *responseValue.Results {
		spew.Dump(result)
		index++
	}

	return
}
