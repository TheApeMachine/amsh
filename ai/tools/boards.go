package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/theapemachine/amsh/utils"
)

type Boards struct {
	client *azuredevops.Connection
}

/*
NewBoards initializes the Boards struct with a connection client.
*/
func NewBoards() (*Boards, error) {
	client := azuredevops.NewPatConnection(os.Getenv("AZURE_DEVOPS_ORG"), os.Getenv("AZURE_DEVOPS_PAT"))
	return &Boards{client: client}, nil
}

/*
Use is the entry point for all operations in Boards tool.
*/
func (boards *Boards) Use(args map[string]any) string {
	if operation, ok := args["operation"].(string); ok {
		switch operation {
		case "wiql":
			if query, ok := args["query"].(string); ok {
				return boards.search(query)
			}
		case "details":
			if id, ok := args["id"].(int); ok {
				return boards.getDetails(id)
			}
		case "create":
			if title, ok := args["title"].(string); ok {
				if description, ok := args["description"].(string); ok {
					if workItemType, ok := args["workItemType"].(string); ok {
						if tags, ok := args["tags"].(string); ok {
							return boards.create(title, description, workItemType, tags)
						}
					}
				}
			}
		case "update":
			if id, ok := args["id"].(int); ok {
				if title, ok := args["title"].(string); ok {
					if description, ok := args["description"].(string); ok {
						if tags, ok := args["tags"].(string); ok {
							return boards.update(id, title, description, tags)
						}
					}
				}
			}
		case "comment":
			// Publish a new comment/reply to a work item
			if id, ok := args["id"].(int); ok {
				if text, ok := args["text"].(string); ok {
					return boards.publishComment(id, text)
				}
			}
		default:
			return "Invalid operation provided."
		}
	}

	return "Invalid arguments provided."
}

/*
create creates a new work item.
*/
func (boards *Boards) create(title, description, workItemType, tags string) string {
	ctx := context.Background()

	// Create work item tracking client
	client, err := workitemtracking.NewClient(ctx, boards.client)
	if err != nil {
		return fmt.Sprintf("Error creating work item tracking client: %v", err)
	}

	// Define the project name or ID
	project := utils.StringPtr("your_project_name") // Replace with your actual project name or ID

	// Define the fields for the new work item
	fields := map[string]interface{}{
		"System.Title":       title,
		"System.Description": description,
		"System.Tags":        tags,
	}

	// Convert fields map to a slice of JsonPatchOperation
	var document []webapi.JsonPatchOperation
	for key, value := range fields {
		document = append(document, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			Path:  utils.StringPtr(fmt.Sprintf("/fields/%s", key)),
			Value: value,
		})
	}

	// Create the work item
	workItem, err := client.CreateWorkItem(ctx, workitemtracking.CreateWorkItemArgs{
		Project:  project,
		Type:     &workItemType,
		Document: &document,
	})
	if err != nil {
		return fmt.Sprintf("Error creating work item: %v", err)
	}

	return fmt.Sprintf("Work item created successfully with ID: %v", *workItem.Id)
}

/*
update updates an existing work item.
*/
func (boards *Boards) update(id int, title, description, tags string) string {
	ctx := context.Background()

	// Create work item tracking client
	client, err := workitemtracking.NewClient(ctx, boards.client)
	if err != nil {
		return fmt.Sprintf("Error creating work item tracking client: %v", err)
	}

	// Define the fields to update
	fields := map[string]interface{}{
		"System.Title":       title,
		"System.Description": description,
		"System.Tags":        tags,
	}

	// Convert fields map to a slice of JsonPatchOperation
	var document []webapi.JsonPatchOperation
	for key, value := range fields {
		document = append(document, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			Path:  utils.StringPtr(fmt.Sprintf("/fields/%s", key)),
			Value: value,
		})
	}

	// Update the work item
	workItem, err := client.UpdateWorkItem(ctx, workitemtracking.UpdateWorkItemArgs{
		Id:       utils.IntPtr(id),
		Document: &document,
	})
	if err != nil {
		return fmt.Sprintf("Error updating work item: %v", err)
	}

	return fmt.Sprintf("Work item updated successfully with ID: %v", *workItem.Id)
}

/*
getDetails retrieves detailed information about a work item, including
its type, title, description, tags, comments, and linked work items.
*/
func (boards *Boards) getDetails(id int) string {
	ctx := context.Background()

	client, err := workitemtracking.NewClient(ctx, boards.client)
	if err != nil {
		return fmt.Sprintf("Error creating work item tracking client: %v", err)
	}

	workItem, err := client.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
		Id:     utils.IntPtr(id),
		Fields: &[]string{"System.WorkItemType", "System.Title", "System.Description", "System.Tags"},
	})

	if err != nil {
		return fmt.Sprintf("Error retrieving work item details: %v", err)
	}

	fields := *workItem.Fields

	details := []string{
		fmt.Sprintf("ID: %v", workItem.Id),
		fmt.Sprintf("Type: %v", fields["System.WorkItemType"]),
		fmt.Sprintf("Title: %v", fields["System.Title"]),
		fmt.Sprintf("Description: %v", fields["System.Description"]),
		fmt.Sprintf("Tags: %v", fields["System.Tags"]),
	}

	details = append(details, boards.getComments(id)...)
	details = append(details, boards.getLinkedWorkItems(id)...)

	return strings.Join(details, "\n")
}

/*
getComments fetches comments for a given work item ID.
*/
func (boards *Boards) getComments(id int) []string {
	ctx := context.Background()

	client, err := workitemtracking.NewClient(ctx, boards.client)
	if err != nil {
		return []string{fmt.Sprintf("Error creating client for comments: %v", err)}
	}

	comments, err := client.GetComments(ctx, workitemtracking.GetCommentsArgs{WorkItemId: utils.IntPtr(id)})
	if err != nil {
		return []string{fmt.Sprintf("Error retrieving comments: %v", err)}
	}

	var commentDetails []string
	commentDetails = append(commentDetails, "Comments:")
	for _, comment := range *comments.Comments {
		commentDetails = append(commentDetails, fmt.Sprintf("- %s", *comment.Text))
	}
	return commentDetails
}

/*
getLinkedWorkItems fetches related work items (links) for a given work item ID.
*/
func (boards *Boards) getLinkedWorkItems(id int) []string {
	ctx := context.Background()

	client, err := workitemtracking.NewClient(ctx, boards.client)
	if err != nil {
		return []string{fmt.Sprintf("Error creating client for linked items: %v", err)}
	}

	workItem, err := client.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
		Id:     utils.IntPtr(id),
		Expand: &workitemtracking.WorkItemExpandValues.Relations,
	})
	if err != nil {
		return []string{fmt.Sprintf("Error retrieving linked work items: %v", err)}
	}

	var linkedItems []string
	linkedItems = append(linkedItems, "Linked Work Items:")
	for _, relation := range *workItem.Relations {
		if *relation.Rel == "System.LinkTypes.Hierarchy-Reverse" || *relation.Rel == "System.LinkTypes.Hierarchy-Forward" {
			linkedItems = append(linkedItems, fmt.Sprintf("- %s", *relation.Url))
		}
	}
	return linkedItems
}

/*
search executes a WIQL query and retrieves work item details.
*/
func (boards *Boards) search(wiql string) string {
	ctx := context.Background()

	// Create work item tracking client
	client, err := workitemtracking.NewClient(ctx, boards.client)
	if err != nil {
		return fmt.Sprintf("Error creating work item tracking client: %v", err)
	}

	// Run WIQL query
	query := workitemtracking.Wiql{
		Query: &wiql,
	}
	queryResult, err := client.QueryByWiql(ctx, workitemtracking.QueryByWiqlArgs{
		Wiql:          &query,
		Project:       utils.StringPtr("fanapp"),
		TimePrecision: utils.BoolPtr(true),
		Top:           utils.IntPtr(20),
	})
	if err != nil {
		return fmt.Sprintf("Error executing WIQL query: %v", err)
	}

	// Gather detailed info for each work item returned by the query
	var details []string
	for _, item := range *queryResult.WorkItems {
		workItem, err := client.GetWorkItem(ctx, workitemtracking.GetWorkItemArgs{
			Id:     item.Id,
			Fields: &[]string{"System.WorkItemType", "System.Title", "System.Description", "System.Tags"},
		})
		if err != nil {
			return fmt.Sprintf("Error retrieving work item details: %v", err)
		}

		fields := *workItem.Fields

		details = append(details, strings.Join([]string{
			fmt.Sprintf("ID: %v", workItem.Id),
			fmt.Sprintf("Type: %v", fields["System.WorkItemType"]),
			fmt.Sprintf("Title: %v", fields["System.Title"]),
			fmt.Sprintf("Description: %v", fields["System.Description"]),
			fmt.Sprintf("Tags: %v", fields["System.Tags"]),
		}, "\n"))
	}

	return strings.Join(details, "\n---\n")
}

/*
publishComment creates a new comment on a work item.
*/
func (boards *Boards) publishComment(id int, text string) string {
	ctx := context.Background()

	// Create work item tracking client
	client, err := workitemtracking.NewClient(ctx, boards.client)
	if err != nil {
		return fmt.Sprintf("Error creating work item tracking client: %v", err)
	}

	// Create a new comment
	comment := workitemtracking.CommentCreate{
		Text: &text,
	}

	// Define the project name or ID
	project := utils.StringPtr("your_project_name") // Replace with your actual project name or ID

	// Publish the comment to the specified work item
	_, err = client.AddComment(ctx, workitemtracking.AddCommentArgs{
		Request:    &comment,
		Project:    project,
		WorkItemId: utils.IntPtr(id),
	})
	if err != nil {
		return fmt.Sprintf("Error publishing comment: %v", err)
	}

	return "Comment published successfully."
}
