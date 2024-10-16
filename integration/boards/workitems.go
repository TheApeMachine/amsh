package boards

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

func (srv *Service) GetWorkitem(ctx context.Context, id int) (out string, err error) {
	var (
		client        workitemtracking.Client
		responseValue *workitemtracking.WorkItem
	)

	if client, err = workitemtracking.NewClient(ctx, azuredevops.NewPatConnection(srv.orgURL, srv.pat)); err != nil {
		return "", errnie.Error(err)
	}

	if responseValue, err = client.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
		Id: &id,
	}); err != nil {
		return "", errnie.Error(err)
	}

	fields := *responseValue.Fields

	template := viper.GetViper().GetString("integrations.boards.workitem")
	template = strings.ReplaceAll(template, "{id}", strconv.Itoa(*responseValue.Id))
	template = strings.ReplaceAll(template, "{title}", fields["System.Title"].(string))
	template = strings.ReplaceAll(template, "{description}", fields["System.Description"].(string))
	template = strings.ReplaceAll(template, "{comments}", fields["System.History"].(string))

	return template, nil
}

func (srv *Service) CreateWorkitem(ctx context.Context, title, description string) (out string, err error) {
	var (
		client workitemtracking.Client
	)

	if client, err = workitemtracking.NewClient(ctx, azuredevops.NewPatConnection(srv.orgURL, srv.pat)); err != nil {
		return "", errnie.Error(err)
	}

	path := "/fields/System.Title"

	doc := []webapi.JsonPatchOperation{
		{
			Op:    &webapi.OperationValues.Add,
			Path:  &path,
			Value: &title,
		},
	}

	responseValue, err := client.CreateWorkItem(ctx, workitemtracking.CreateWorkItemArgs{
		Project:  &srv.projectName,
		Document: &doc,
	})

	if err != nil {
		return "[ERROR] Failed to create work item [/ERROR]", errnie.Error(err)
	}

	template := viper.GetViper().GetString("integrations.boards.response")
	template = strings.ReplaceAll(template, "{id}", strconv.Itoa(*responseValue.Id))
	template = strings.ReplaceAll(template, "{response}", "Work item created")

	return template, nil
}

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
