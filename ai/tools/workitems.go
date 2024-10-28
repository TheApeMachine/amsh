package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	wi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/types"
	"github.com/theapemachine/amsh/errnie"
)

type WorkItemsTool struct {
	client      wi.Client
	projectName string
}

func (w *WorkItemsTool) Description() string {
	return viper.GetViper().GetString("tools.workitems")
}

func NewWorkItemsTool(ctx context.Context) (*WorkItemsTool, error) {
	organizationURL := os.Getenv("AZDO_ORG_URL")
	personalAccessToken := os.Getenv("AZDO_PAT")
	projectName := os.Getenv("AZDO_PROJECT_NAME")

	if organizationURL == "" || personalAccessToken == "" || projectName == "" {
		return nil, errors.New("azure devops configuration is missing")
	}

	connection := azuredevops.NewPatConnection(organizationURL, personalAccessToken)
	client, err := wi.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure DevOps client: %w", err)
	}

	return &WorkItemsTool{
		client:      client,
		projectName: projectName,
	}, nil
}

func (w *WorkItemsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	operation, err := getStringArg(args, "operation", "")
	if err != nil || operation == "" {
		return "", errors.New("operation is required")
	}

	switch operation {
	case "fetch":
		return w.fetchWorkItem(ctx, args)
	case "create":
		return w.createWorkItem(ctx, args)
	case "update":
		return w.updateWorkItem(ctx, args)
	default:
		return "", fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (w *WorkItemsTool) fetchWorkItem(ctx context.Context, args map[string]interface{}) (string, error) {
	idStr, err := getStringArg(args, "id", "")
	if err != nil || idStr == "" {
		return "", errors.New("id is required for fetch operation")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return "", errors.New("invalid id format")
	}

	workItem, err := w.client.GetWorkItem(ctx, wi.GetWorkItemArgs{
		Id: &id,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch work item: %w", err)
	}

	fields := workItem.Fields
	title := getFieldValue(*fields, "System.Title")
	description := getFieldValue(*fields, "System.Description")
	state := getFieldValue(*fields, "System.State")

	output := fmt.Sprintf("Work Item ID: %d\nTitle: %s\nDescription: %s\nState: %s\n", id, title, description, state)
	return output, nil
}

func (w *WorkItemsTool) createWorkItem(ctx context.Context, args map[string]interface{}) (string, error) {
	title, err := getStringArg(args, "title", "")
	if err != nil || title == "" {
		return "", errors.New("title is required for create operation")
	}

	titlePath := "/fields/System.Title"
	descriptionPath := "/fields/System.Description"
	description, err := getStringArg(args, "description", "")
	if err != nil {
		return "", errors.New("description is required for create operation")
	}

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

	parentID, err := getIntArg(args, "parent_id", nil)
	if err != nil {
		return "", errors.New("parent_id is required for create operation")
	}

	if parentID != 0 {
		// Add a link to the parent work item
		parentLinkPath := "/relations/-"
		parentLink := map[string]interface{}{
			"rel": "System.LinkTypes.Hierarchy-Reverse",
			"url": fmt.Sprintf("https://dev.azure.com/{organization}/{project}/_apis/wit/workItems/%d", parentID),
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

	workitemType, err := getStringArg(args, "workitem_type", "Task")
	if err != nil {
		return "", errors.New("workitem_type is required for create operation")
	}

	responseValue, err := w.client.CreateWorkItem(ctx, wi.CreateWorkItemArgs{
		Project:  &w.projectName,
		Document: &doc,
		Type:     &workitemType,
	})

	if err != nil {
		return "", errnie.Error(err)
	}

	output := fmt.Sprintf("Created Work Item ID: %d\n", *responseValue.Id)
	return output, nil
}

func (w *WorkItemsTool) updateWorkItem(ctx context.Context, args map[string]interface{}) (string, error) {
	idStr, err := getStringArg(args, "id", "")
	if err != nil || idStr == "" {
		return "", errors.New("id is required for update operation")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return "", errors.New("invalid id format")
	}

	title, _ := getStringArg(args, "title", "")
	description, _ := getStringArg(args, "description", "")

	titlePath := "/fields/System.Title"
	descriptionPath := "/fields/System.Description"

	var document []webapi.JsonPatchOperation

	if title != "" {
		document = append(document, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			Path:  &titlePath,
			Value: &title,
		})
	}
	if description != "" {
		document = append(document, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			Path:  &descriptionPath,
			Value: &description,
		})
	}

	if len(document) == 0 {
		return "", errors.New("no fields to update")
	}

	workItem, err := w.client.UpdateWorkItem(ctx, wi.UpdateWorkItemArgs{
		Id:       &id,
		Document: &document,
	})
	if err != nil {
		return "", fmt.Errorf("failed to update work item: %w", err)
	}

	output := fmt.Sprintf("Updated Work Item ID: %d\n", *workItem.Id)
	return output, nil
}

func (w *WorkItemsTool) GetSchema() types.ToolSchema {
	return types.ToolSchema{
		Name:        "work_items",
		Description: "Interact with Azure DevOps work items.",
		Parameters: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The operation to perform on the work items.",
				"enum":        []string{"fetch", "create", "update"},
			},
			"id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the work item to fetch or update.",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "The title of the work item to create or update.",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "The description of the work item.",
			},
			"workitem_type": map[string]interface{}{
				"type":        "string",
				"description": "The type of the work item to create (e.g., Task, Bug).",
			},
		},
	}
}

func getFieldValue(fields map[string]interface{}, field string) string {
	if value, ok := fields[field]; ok {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

func getIntArg(args map[string]interface{}, key string, defaultValue *int) (int, error) {
	if value, ok := args[key]; ok {
		return cast.ToInt(value), nil
	}
	return *defaultValue, nil
}
