package boards

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/search"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/spf13/viper"
	"github.com/theapemachine/errnie"
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
	ctx context.Context, title, description, workitemType string, parentID *int,
) (out string, err error) {
	var (
		responseValue *workitemtracking.WorkItem
	)

	titlePath := "/fields/System.Title"
	descriptionPath := "/fields/System.Description"

	doc := []webapi.JsonPatchOperation{
		{
			Op:    &webapi.OperationValues.Add,
			Path:  &titlePath,
			Value: &title,
		},
		{
			Op:    &webapi.OperationValues.Add,
			Path:  &descriptionPath,
			Value: &description,
		},
	}

	if parentID != nil {
		// Add a link to the parent work item
		parentLinkPath := "/relations/-"
		parentLink := map[string]interface{}{
			"rel": "System.LinkTypes.Hierarchy-Reverse",
			"url": fmt.Sprintf("https://dev.azure.com/{organization}/{project}/_apis/wit/workItems/%d", *parentID),
			"attributes": map[string]string{
				"comment": "Linking to parent work item",
			},
		}
		doc = append(doc, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			Path:  &parentLinkPath,
			Value: &parentLink,
		})
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
		filters       = map[string][]string{}
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

		builder.WriteString("----------------------------------------\n")
		builder.WriteString(fmt.Sprintf("Id: %s\n", tickets["System.Id"]))
		builder.WriteString(fmt.Sprintf("Type: %s\n", tickets["System.WorkItemType"]))
		builder.WriteString(fmt.Sprintf("State: %s\n", tickets["System.State"]))
		builder.WriteString(fmt.Sprintf("Assigned To: %s\n", tickets["System.AssignedTo"]))
		builder.WriteString(fmt.Sprintf("Created Date: %s\n", tickets["System.CreatedDate"]))
		builder.WriteString(fmt.Sprintf("Tags: %s\n", tickets["System.Tags"]))
		builder.WriteString(fmt.Sprintf("Title: %s\n", tickets["System.Title"]))
		builder.WriteString(fmt.Sprintf("Description: %s\n", tickets["System.Description"]))

		// Get comments for this work item
		if id, ok := tickets["System.Id"]; ok {
			workItemID, _ := strconv.Atoi(fmt.Sprintf("%v", id))
			commentSrv, err := NewGetCommentsSrv(ctx, os.Getenv("AZDO_ORG_URL"), os.Getenv("AZDO_PAT"))
			if err == nil {
				if comments, err := commentSrv.GetComments(ctx, workItemID, nil); err == nil {
					builder.WriteString(fmt.Sprintf("Comments:\n%s\n", comments))
				}
			}
		}
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

type GetCommentsSrv struct {
	conn workitemtracking.Client
}

func NewGetCommentsSrv(ctx context.Context, orgURL, pat string) (*GetCommentsSrv, error) {
	conn, err := workitemtracking.NewClient(ctx, azuredevops.NewPatConnection(orgURL, pat))
	if err != nil {
		return nil, errnie.Error(err)
	}

	return &GetCommentsSrv{conn: conn}, nil
}

func (srv *GetCommentsSrv) GetComments(ctx context.Context, workItemId int, commentIds []int) (out string, err error) {
	var (
		responseValue *workitemtracking.CommentList
	)

	if responseValue, err = srv.conn.GetCommentsBatch(ctx, workitemtracking.GetCommentsBatchArgs{
		WorkItemId: &workItemId,
		Ids:        &commentIds, // Pass commentIds directly
	}); err != nil {
		return "", errnie.Error(err)
	}

	builder := strings.Builder{}
	for _, comment := range *responseValue.Comments { // Added * to dereference the slice
		builder.WriteString(fmt.Sprintf("Comment ID: %d\n", *comment.Id))
		builder.WriteString(fmt.Sprintf("Text: %s\n", *comment.Text)) // Added * to dereference
		builder.WriteString(fmt.Sprintf("Created By: %s\n", *comment.CreatedBy.DisplayName))
		builder.WriteString(fmt.Sprintf("Created Date: %s\n", comment.CreatedDate))
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
