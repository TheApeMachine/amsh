package boards

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/theapemachine/amsh/errnie"
)

func (srv *Service) SearchWorkitems(ctx context.Context, query string) (out string, err error) {
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
		return "", errnie.Error(err)
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
		return "", errnie.Error(err)
	}

	if responseValue.Results == nil {
		log.Println("No results found")
		return
	}

	builder := strings.Builder{}
	for _, result := range *responseValue.Results {
		tickets := *result.Fields

		spew.Dump(tickets)
		builder.WriteString(fmt.Sprintf("%s\n", tickets["system.workitemtype"]))
		builder.WriteString(fmt.Sprintf("%s\n", tickets["system.title"]))
		builder.WriteString(fmt.Sprintf("%s\n", tickets["system.description"]))
		builder.WriteString(fmt.Sprintf("%s\n", tickets["system.state"]))
		builder.WriteString(fmt.Sprintf("%s\n", tickets["system.assignedto"]))
		builder.WriteString(fmt.Sprintf("%s\n", tickets["system.createddate"]))
		builder.WriteString(fmt.Sprintf("%s\n", tickets["system.tags"]))
		builder.WriteString(fmt.Sprintf("%s\n", tickets["---"]))
		index++
	}

	return builder.String(), nil
}
