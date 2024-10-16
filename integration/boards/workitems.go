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
	"github.com/openai/openai-go"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai"
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

func (srv *GetWorkItemSrv) Call(args map[string]any) (string, error) {
	return "", nil
}

func (srv *GetWorkItemSrv) Schema() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"get_workitem",
		"Get a work item from the board",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]string{
					"type":        "integer",
					"description": "The ID of the work item to get",
				},
			},
			"required": []string{"id"},
		},
	)
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

	return &CreateWorkitemSrv{conn: conn}, nil
}

func (srv *CreateWorkitemSrv) CreateWorkitem(
	ctx context.Context, title, description string,
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
	})

	if err != nil {
		return "", errnie.Error(err)
	}

	template := viper.GetViper().GetString("integrations.boards.response")
	template = strings.ReplaceAll(template, "{id}", strconv.Itoa(*responseValue.Id))
	template = strings.ReplaceAll(template, "{response}", "Work item created")

	return template, nil
}

func (srv *CreateWorkitemSrv) Call(args map[string]any) (string, error) {
	return "", nil
}

func (srv *CreateWorkitemSrv) Schema() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"create_workitem",
		"Create a work item on the board",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]string{
					"type":        "string",
					"description": "The title of the work item to create",
				},
				"description": map[string]string{
					"type":        "string",
					"description": "The description of the work item to create",
				},
			},
			"required": []string{"title", "description"},
		},
	)
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

func (srv *SearchWorkitemsSrv) Call(args map[string]any) (string, error) {
	return "", nil
}

func (srv *SearchWorkitemsSrv) Schema() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"search_workitems",
		"Search for work items on the board",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]string{
					"type":        "string",
					"description": "The query to search for",
				},
			},
			"required": []string{"query"},
		},
	)
}
