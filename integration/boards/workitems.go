package boards

import (
	"context"
	"fmt"
	"log"
	"os"
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

type GetWorkItemSrv struct {
	conn workitemtracking.Client
}

func NewGetWorkItemSrv(ctx context.Context, orgURL, pat string) (*GetWorkItemSrv, error) {
	conn, err := workitemtracking.NewClient(ctx, azuredevops.NewPatConnection(orgURL, pat))
	if err != nil {
		return nil, errnie.Error(err)
	}

	return &GetWorkItemSrv{conn: conn}, nil
}

func (srv *GetWorkItemSrv) GetWorkitem(ctx context.Context, id int) (out string, err error) {
	var (
		responseValue *workitemtracking.WorkItem
	)

	if responseValue, err = srv.conn.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
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

type CreateWorkitemSrv struct {
	conn        workitemtracking.Client
	projectName string
}

func NewCreateWorkitemSrv(ctx context.Context, projectName string) (*CreateWorkitemSrv, error) {
	conn, err := workitemtracking.NewClient(ctx, azuredevops.NewPatConnection(
		os.Getenv("AZDO_ORG_URL"),
		os.Getenv("AZDO_PAT"),
	))
	if err != nil {
		return nil, errnie.Error(err)
	}

	return &CreateWorkitemSrv{conn: conn, projectName: "playground"}, nil
}

func (srv *CreateWorkitemSrv) CreateWorkitem(
	ctx context.Context, title, description, workitemType string,
) (out string, err error) {
	var (
		responseValue *workitemtracking.WorkItem
	)

	path := "/fields/System.Title"
	title = "test"

	doc := []webapi.JsonPatchOperation{
		{
			Op:    &webapi.OperationValues.Add,
			Path:  &path,
			Value: &title,
		},
	}

	responseValue, err = srv.conn.CreateWorkItem(ctx, workitemtracking.CreateWorkItemArgs{
		Project:  &srv.projectName,
		Document: &doc,
		Type:     &workitemType,
	})

	if err != nil {
		return "", errnie.Error(err)
	}

	template := viper.GetViper().GetString("integrations.boards.response")
	template = strings.ReplaceAll(template, "{id}", strconv.Itoa(*responseValue.Id))
	template = strings.ReplaceAll(template, "{response}", "Work item created")

	return template, nil
}

type SearchWorkitemsSrv struct {
	conn        search.Client
	projectName string
}

func NewSearchWorkitemsSrv(ctx context.Context, projectName string) (*SearchWorkitemsSrv, error) {
	conn, err := search.NewClient(ctx, azuredevops.NewPatConnection(
		os.Getenv("AZDO_ORG_URL"),
		os.Getenv("AZDO_PAT"),
	))
	if err != nil {
		return nil, errnie.Error(err)
	}

	return &SearchWorkitemsSrv{conn: conn}, nil
}

func (srv *SearchWorkitemsSrv) SearchWorkitems(ctx context.Context, query string) (out string, err error) {
	var (
		responseValue *search.WorkItemSearchResponse
		index         = 0
		skip          = 0
		top           = 10
		filters       = map[string][]string{
			"System.State": {"To Do"},
		}
	)

	if responseValue, err = srv.conn.FetchWorkItemSearchResults(ctx, search.FetchWorkItemSearchResultsArgs{
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

type UpdateWorkitemSrv struct {
	conn        workitemtracking.Client
	projectName string
}

func NewUpdateWorkitemSrv(ctx context.Context, projectName string) (*UpdateWorkitemSrv, error) {
	conn, err := workitemtracking.NewClient(ctx, azuredevops.NewPatConnection(
		os.Getenv("AZDO_ORG_URL"),
		os.Getenv("AZDO_PAT"),
	))
	if err != nil {
		return nil, errnie.Error(err)
	}

	return &UpdateWorkitemSrv{conn: conn, projectName: projectName}, nil
}

func (srv *UpdateWorkitemSrv) UpdateWorkitem(ctx context.Context, id int, title, description string) (out string, err error) {
	var responseValue *workitemtracking.WorkItem

	path := "/fields/System.Title"

	document := []webapi.JsonPatchOperation{
		{
			Op:    &webapi.OperationValues.Add,
			Path:  &path,
			Value: &title,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  &path,
			Value: &description,
		},
	}

	responseValue, err = srv.conn.UpdateWorkItem(ctx, workitemtracking.UpdateWorkItemArgs{
		Id:       &id,
		Project:  &srv.projectName,
		Document: &document,
	})

	if err != nil {
		return "", errnie.Error(err)
	}

	template := viper.GetViper().GetString("integrations.boards.response")
	template = strings.ReplaceAll(template, "{id}", strconv.Itoa(*responseValue.Id))
	template = strings.ReplaceAll(template, "{response}", "Work item updated")

	return template, nil
}
