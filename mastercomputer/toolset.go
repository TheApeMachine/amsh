package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/integration/boards"
	"github.com/theapemachine/amsh/integration/comms"
	"github.com/theapemachine/amsh/integration/git"
	"github.com/theapemachine/amsh/integration/trengo"
	"github.com/theapemachine/amsh/utils"
	"github.com/tmc/langchaingo/schema"
)

var toolsetInstance *Toolset
var toolsetOnce sync.Once

// Toolset represents a set of tools available to a worker.
type Toolset struct {
	toolMap   map[string]openai.ChatCompletionToolParam
	baseTools []string
	workloads map[string][]string
}

/*
NewToolset returns a toolset based on the workload of the worker.
*/
func NewToolset() *Toolset {
	toolsetOnce.Do(func() {
		toolsetInstance = &Toolset{
			toolMap: make(map[string]openai.ChatCompletionToolParam),
			baseTools: []string{
				"add_vector_memory",
				"search_vector_memory",
				"add_graph_memory",
				"search_graph_memory",
			},
			workloads: map[string][]string{
				"discussion": {
					"change_state",
				},
				"sequencer": {
					"assignment",
				},
				"researcher": {
					"search_github_code",
					"search_slack_messages",
					"search_helpdesk_tickets",
					"search_work_items",
					"get_helpdesk_messages",
					"get_helpdesk_message",
					"use_browser",
				},
				"actor": {
					"start_environment",
					"search_github_code",
					"add_helpdesk_labels",
					"get_helpdesk_labels",
				},
				"planner": {
					"search_slack_messages",
					"search_helpdesk_tickets",
					"search_work_items",
					"get_work_item",
					"manage_work_item",
				},
			},
		}
		toolsetInstance.makeTools()
	})

	return toolsetInstance
}

/*
Assign a set of tools to a worker based on the workload.
*/
func (toolset *Toolset) Assign(workload string) (out []openai.ChatCompletionToolParam) {
	errnie.Trace()

	for _, tool := range toolset.baseTools {
		if tool, ok := toolset.toolMap[tool]; ok {
			out = append(out, tool)
		}
	}

	if tools, ok := toolset.workloads[workload]; ok {
		for _, tool := range tools {
			if tool, ok := toolset.toolMap[tool]; ok {
				out = append(out, tool)
			}
		}
	}

	return out
}

func (toolset *Toolset) getArgPresent(args map[string]any, arg string) (out bool) {
	errnie.Trace()

	if _, ok := args[arg]; ok {
		return true
	}
	return false
}

/*
Use a tool, based on the tool call passed in.
*/
func (toolset *Toolset) Use(sequencer *Sequencer, worker *Worker, toolCall openai.ChatCompletionMessageToolCall) (out openai.ChatCompletionToolMessageParam) {
	errnie.Trace()

	args := map[string]any{}
	var err error

	if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return openai.ToolMessage(toolCall.ID, err.Error())
	}

	switch toolCall.Function.Name {
	case "change_state":
		if toolset.getArgPresent(args, "state") {
			var state string
			var ok bool
			if state, ok = args["state"].(string); ok && state == "agreed" {
				worker.state = WorkerStateAgreed
			} else if state, ok = args["state"].(string); ok && state == "disagreed" {
				worker.state = WorkerStateDisagreed
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Worker state changed to %s", state))
		}
	case "assignment":
		if toolset.getArgPresent(args, "role") && toolset.getArgPresent(args, "assignment") {
			role := args["role"].(string)
			assignment := args["assignment"].(string)
			workers := sequencer.workers[role]
			discussion := NewExecutor(sequencer)
			discussion.conversation.context = append([]openai.ChatCompletionMessageParamUnion{}, sequencer.executor.conversation.context...)
			discussion.conversation.Update(openai.AssistantMessage("[NEXT ASSIGNMENT]\n" + assignment + "\n[/NEXT ASSIGNMENT]\n"))
			discussion.conversation.Update(openai.SystemMessage("Since you are a team of 3, all with the same role, discuss how to divide the work for the assignment."))

			for _, wrkr := range workers {
				wrkr.discussion = discussion
				wrkr.state = WorkerStateDiscussing
				wrkr.Start()
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Assigned '%s' to workers with role '%s'", assignment, role))
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for assignment")
	case "worker":
		if toolset.getArgPresent(args, "system_prompt") && toolset.getArgPresent(args, "user_prompt") && toolset.getArgPresent(args, "toolset") {
			worker := NewWorker(sequencer.ctx, utils.NewName(), toolset.Assign(args["toolset"].(string)), sequencer.executor, worker.role)
			worker.system = args["system_prompt"].(string)
			worker.user = args["user_prompt"].(string)
			worker.Initialize()
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("worker (%s) created", worker.name))
		}

		return openai.ToolMessage(toolCall.ID, "something went wrong")
	case "add_vector_memory":
		if toolset.getArgPresent(args, "content") && toolset.getArgPresent(args, "private") {
			content := args["content"].(string)
			store := memory.NewQdrant("hive", 1536)

			if err = store.AddDocuments([]schema.Document{
				{
					PageContent: content,
					Metadata: map[string]any{
						"worker": sequencer.worker.name,
						"user":   sequencer.worker.user,
					},
				},
			}); err != nil {
				return openai.ToolMessage(toolCall.ID, "Error adding vector memory: "+err.Error())
			}

			return openai.ToolMessage(toolCall.ID, "vector memory added")
		}

		return openai.ToolMessage(toolCall.ID, "Invalid arguments for add_vector_memory")
	case "search_vector_memory":
		if toolset.getArgPresent(args, "content") && toolset.getArgPresent(args, "private") {
			content := args["content"].(string)
			store := memory.NewQdrant("hive", 1536)

			// Implement the logic to search vector memory
			result, err := store.Query(fmt.Sprintf("SEARCH %s %v", content, args["private"].(bool)))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching vector memory: "+err.Error())
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Vector memory search results: %v", result))
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_vector_memory")

	case "use_graph_memory":
		if toolset.getArgPresent(args, "cypher") {
			store := memory.NewNeo4j()
			operation := args["operation"].(string)
			cypher := args["cypher"].(string)

			// Implement the logic to add graph memory
			if operation == "add" {
				result, err := store.Write(cypher)
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error adding graph memory: "+err.Error())
				}
				return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Graph memory added: %v", result))
			}

			if operation == "update" {
				result, err := store.Write(cypher)
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error updating graph memory: "+err.Error())
				}
				return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Graph memory updated: %v", result))
			}

			if operation == "search" {
				result, err := store.Query(cypher)
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error searching graph memory: "+err.Error())
				}
				return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Graph memory search results: %v", result))
			}

			return openai.ToolMessage(toolCall.ID, "Invalid arguments for use_graph_memory")
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for add_graph_memory")
	case "search_work_items":
		if toolset.getArgPresent(args, "query") {
			query := args["query"].(string)
			ctx := context.Background()
			searchSrv, err := boards.NewSearchWorkitemsSrv(ctx, "your_project_name")
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error creating search service: "+err.Error())
			}
			result, err := searchSrv.SearchWorkitems(ctx, query)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching work items: "+err.Error())
			}
			return openai.ToolMessage(toolCall.ID, result)
		}

		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_work_items")

	case "get_work_item":
		if toolset.getArgPresent(args, "id") {
			id, err := strconv.Atoi(args["id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid work item ID: "+err.Error())
			}
			ctx := context.Background()
			getSrv, err := boards.NewGetWorkItemSrv(ctx, os.Getenv("AZDO_ORG_URL"), os.Getenv("AZDO_PAT"))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error creating get service: "+err.Error())
			}
			result, err := getSrv.GetWorkitem(ctx, id)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error getting work item: "+err.Error())
			}
			return openai.ToolMessage(toolCall.ID, result)
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for get_work_item")

	case "manage_work_item":
		if toolset.getArgPresent(args, "action") && toolset.getArgPresent(args, "title") && toolset.getArgPresent(args, "description") {
			action := args["action"].(string)
			title := args["title"].(string)
			description := args["description"].(string)
			workItemType := args["workitem_type"].(string)
			ctx := context.Background()

			var parentID *int
			if toolset.getArgPresent(args, "parent_id") {
				id, err := strconv.Atoi(args["parent_id"].(string))
				if err == nil {
					parentID = &id
				}
			}

			switch action {
			case "create":
				createSrv, err := boards.NewCreateWorkitemSrv(ctx, "playground")
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error creating create service: "+err.Error())
				}
				result, err := createSrv.CreateWorkitem(ctx, title, description, workItemType, parentID)
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error creating work item: "+err.Error())
				}
				return openai.ToolMessage(toolCall.ID, result)
			case "update":
				if !toolset.getArgPresent(args, "id") {
					return openai.ToolMessage(toolCall.ID, "ID is required for updating a work item")
				}
				id, err := strconv.Atoi(args["id"].(string))
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Invalid work item ID: "+err.Error())
				}
				updateSrv, err := boards.NewUpdateWorkitemSrv(ctx, "playground")
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error creating update service: "+err.Error())
				}
				result, err := updateSrv.UpdateWorkitem(ctx, id, title, description)
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error updating work item: "+err.Error())
				}
				return openai.ToolMessage(toolCall.ID, result)
			default:
				return openai.ToolMessage(toolCall.ID, "Invalid action for manage_work_item")
			}
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for manage_work_item")

	case "search_helpdesk_tickets":
		if toolset.getArgPresent(args, "query") && toolset.getArgPresent(args, "page") {
			query := args["query"].(string)
			pageFloat, ok := args["page"].(float64) // Safely assert to float64
			if !ok {
				return openai.ToolMessage(toolCall.ID, "Invalid page number type")
			}
			page := int(pageFloat) // Convert float64 to int

			ticketService := trengo.NewTicketService()
			tickets, err := ticketService.ListTickets(context.Background(), page)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching helpdesk tickets: "+err.Error())
			}

			// Format the results for the AI
			formattedResults := fmt.Sprintf("Helpdesk Tickets (Page %d, Query: %s):\n\n", page, query)
			for _, ticket := range tickets {
				formattedResults += fmt.Sprintf("ID: %d\nSubject: %s\nStatus: %s\nCreated At: %s\nUpdated At: %s\nAssigned To: %s\nLast Message: %s\n\n",
					ticket.ID, ticket.Subject, ticket.Status, ticket.CreatedAt, ticket.UpdatedAt, ticket.AssignedTo, ticket.LastMessage)
			}

			return openai.ToolMessage(toolCall.ID, formattedResults)
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_helpdesk_tickets")

	case "get_helpdesk_messages":
		if toolset.getArgPresent(args, "ticket_id") {
			// Convert float64 to int
			ticketIDFloat, ok := args["ticket_id"].(float64)
			if !ok {
				return openai.ToolMessage(toolCall.ID, "Invalid ticket ID type")
			}
			ticketID := int(ticketIDFloat)

			messageService := trengo.NewMessageService(os.Getenv("TRENGO_BASE_URL"), os.Getenv("TRENGO_API_TOKEN"))
			messages, err := messageService.ListMessages(context.Background(), ticketID)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error getting helpdesk messages: "+err.Error())
			}

			// Format the results for the AI
			formattedResults := fmt.Sprintf("Helpdesk Messages for Ticket %d:\n\n", ticketID)
			for _, message := range messages {
				formattedResults += fmt.Sprintf("ID: %d\nBody: %s\nCreated At: %s\nUser: %s (ID: %d)\n\n",
					message.ID, message.Body, message.CreatedAt, message.User.Name, message.User.ID)
			}

			return openai.ToolMessage(toolCall.ID, formattedResults)
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for get_helpdesk_messages")

	case "get_helpdesk_message":
		if toolset.getArgPresent(args, "ticket_id") && toolset.getArgPresent(args, "message_id") {
			ticketID, err := strconv.Atoi(args["ticket_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid ticket ID: "+err.Error())
			}
			messageID, err := strconv.Atoi(args["message_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid message ID: "+err.Error())
			}

			messageService := trengo.NewMessageService(os.Getenv("TRENGO_BASE_URL"), os.Getenv("TRENGO_API_TOKEN"))
			message, err := messageService.FetchMessage(context.Background(), ticketID, messageID)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error getting helpdesk message: "+err.Error())
			}

			// Format the result for the AI
			formattedResult := fmt.Sprintf("Helpdesk Message (Ticket %d, Message %d):\n\nBody: %s\nCreated At: %s\nUser: %s (ID: %d)",
				ticketID, messageID, message.Body, message.CreatedAt, message.User.Name, message.User.ID)

			return openai.ToolMessage(toolCall.ID, formattedResult)
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for get_helpdesk_message")

	case "get_helpdesk_labels":
		ctx := context.Background()
		listLabels := trengo.NewListLabels()
		labels, err := listLabels.List(ctx)
		if err != nil {
			return openai.ToolMessage(toolCall.ID, "Error getting helpdesk labels: "+err.Error())
		}
		labelsJSON, err := json.Marshal(labels)
		if err != nil {
			return openai.ToolMessage(toolCall.ID, "Error marshaling labels: "+err.Error())
		}
		return openai.ToolMessage(toolCall.ID, string(labelsJSON))
	case "add_helpdesk_labels":
		if toolset.getArgPresent(args, "ticket_id") && toolset.getArgPresent(args, "label_id") {
			ticketID, err := strconv.Atoi(args["ticket_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid ticket ID: "+err.Error())
			}
			labelID, err := strconv.Atoi(args["label_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid label ID: "+err.Error())
			}
			ctx := context.Background()
			assignLabels := trengo.NewAssignLabels()
			err = assignLabels.Attach(ctx, labelID, ticketID)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error adding helpdesk label: "+err.Error())
			}
			return openai.ToolMessage(toolCall.ID, "Label successfully added to the ticket")
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for add_helpdesk_labels")
	case "search_slack_messages":
		if toolset.getArgPresent(args, "query") {
			search := comms.NewSearch()
			messages, err := search.SearchMessages(context.Background(), args["query"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching Slack messages: "+err.Error())
			}
			result, _ := json.Marshal(messages)
			return openai.ToolMessage(toolCall.ID, string(result))
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_slack_messages")
	case "send_slack_channel_message":
		if toolset.getArgPresent(args, "channel") && toolset.getArgPresent(args, "message") {
			msg := comms.NewMessage(args["channel"].(string), args["message"].(string))
			err := msg.Post(context.Background())
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error sending Slack channel message: "+err.Error())
			}
			return openai.ToolMessage(toolCall.ID, "Slack channel message sent successfully")
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for send_slack_channel_message")
	case "send_slack_user_message":
		if toolset.getArgPresent(args, "user") && toolset.getArgPresent(args, "message") {
			// Note: This assumes the user ID is provided. You might need to look up the user ID from the username if that's what's provided.
			msg := comms.NewMessage(args["user"].(string), args["message"].(string))
			err := msg.Post(context.Background())
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error sending Slack user message: "+err.Error())
			}
			return openai.ToolMessage(toolCall.ID, "Slack user message sent successfully")
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for send_slack_user_message")
	case "search_github_code":
		if toolset.getArgPresent(args, "query") {
			hub := git.NewHub()
			results, err := hub.SearchCode(context.Background(), args["query"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching GitHub code: "+err.Error())
			}

			// Format the results for the AI
			formattedResults := "GitHub Code Search Results:\n\n"
			for _, result := range results {
				formattedResults += fmt.Sprintf("Repository: %s\nFile: %s\nContent:\n```\n%s\n```\n\n",
					result.Repository, result.Path, result.Content)
			}

			return openai.ToolMessage(toolCall.ID, formattedResults)
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_github_code")
	case "add_code_review":
		return openai.ToolMessage(toolCall.ID, "tool not implemented")
	case "use_browser":
		if toolset.getArgPresent(args, "url") && toolset.getArgPresent(args, "javascript") {
			browser := NewBrowser()
			result, err := browser.Run(args)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error using browser: "+err.Error())
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("\n[BROWSER RESULT]\n%s\n[/BROWSER RESULT]\n", result))
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for use_browser")
	case "start_environment":
		return openai.ToolMessage(toolCall.ID, "tool not implemented")
	}

	return openai.ToolMessage(toolCall.ID, "Tool not implemented")
}

func (toolset *Toolset) makeTools() {
	toolset.toolMap["change_state"] = toolset.makeSchema(
		"change_state",
		"Change your state when in a discussion. A discussion will go on until all workers have agreed on the decisions.",
		map[string]interface{}{
			"state": toolset.makeEnumParam("The state to change to.", []string{"disagree", "agreed"}),
		},
	)
	toolset.toolMap["assignment"] = toolset.makeSchema(
		"assignment",
		"Assign the next team of workers in your sequence.",
		map[string]interface{}{
			"role":       toolset.makeEnumParam("The role of the worker to assign.", []string{"prompt", "reasoner", "researcher", "planner", "actor"}),
			"assignment": toolset.makeStringParam("The assignment to assign to the worker."),
		},
	)

	toolset.toolMap["worker"] = toolset.makeSchema(
		"worker",
		"Create a new worker for any kind of task. A worker can be given a system prompt and a user prompt, as well as a toolset to use.",
		map[string]interface{}{
			"system_prompt": toolset.makeStringParam("The system prompt for the worker."),
			"user_prompt":   toolset.makeStringParam("The user prompt for the worker."),
			"toolset":       toolset.makeEnumParam("The toolset to use for the worker.", []string{"reasoning", "verifying", "executing"}),
		},
	)

	toolset.toolMap["add_vector_memory"] = toolset.makeSchema(
		"add_vector_memory",
		"Add a memory to the vector store. Useful for storing memories as vectors.",
		map[string]interface{}{
			"content": toolset.makeStringParam("The content of the memory to add to the vector store."),
		},
	)

	toolset.toolMap["search_vector_memory"] = toolset.makeSchema(
		"search_vector_memory",
		"Search the vector store. Useful for retrieving memories from the vector store.",
		map[string]interface{}{
			"question": toolset.makeStringParam("The question to search the vector memory for."),
		},
	)

	toolset.toolMap["use_graph_memory"] = toolset.makeSchema(
		"add_graph_memory",
		"Add a memory to the graph store. Useful for storing memories as connected relationships. Make sure you have at least one relationship defined in your cypher query.",
		map[string]interface{}{
			"operation": toolset.makeEnumParam("The operation to perform on the graph store.", []string{"add", "update", "search"}),
			"cypher":    toolset.makeStringParam("The cypher query to add new nodes and edges to the graph store."),
		},
	)

	toolset.toolMap["search_work_items"] = toolset.makeSchema(
		"search_work_items",
		"Search the work items. Useful for retrieving work items from the project management system.",
		map[string]interface{}{
			"query": toolset.makeStringParam("The query to search the work items for."),
		},
	)

	toolset.toolMap["get_work_item"] = toolset.makeSchema(
		"get_work_item",
		"Get a work item. Useful for retrieving a work item from the project management system.",
		map[string]interface{}{
			"id": toolset.makeStringParam("The id of the work item you want to get."),
		},
	)

	toolset.toolMap["manage_work_item"] = toolset.makeSchema(
		"manage_work_item",
		"Manage a work item. Useful for creating or updating work items in the project management system.",
		map[string]interface{}{
			"action":        toolset.makeEnumParam("The action to perform on the work item.", []string{"create", "update"}),
			"workitem_type": toolset.makeEnumParam("The type of work item to create.", []string{"Epic", "Issue", "Task"}),
			"title":         toolset.makeStringParam("The title of the work item."),
			"description":   toolset.makeStringParam("Gherkin description of the work item."),
		},
	)

	toolset.toolMap["search_helpdesk_tickets"] = toolset.makeSchema(
		"search_helpdesk_tickets",
		"Search for helpdesk tickets in Trengo.",
		map[string]interface{}{
			"query": toolset.makeStringParam("The search query to use."),
			"page":  toolset.makeIntParam("The page number of results to retrieve."),
		},
	)

	toolset.toolMap["get_helpdesk_messages"] = toolset.makeSchema(
		"get_helpdesk_messages",
		"Get messages for a specific helpdesk ticket in Trengo.",
		map[string]interface{}{
			"ticket_id": toolset.makeIntParam("The ID of the ticket to get messages for."),
		},
	)

	toolset.toolMap["add_helpdesk_labels"] = toolset.makeSchema(
		"add_helpdesk_labels",
		"Add category and keyword labels to a helpdesk ticket. Very useful and highly appreciated by the support team.",
		map[string]interface{}{
			"id":     toolset.makeStringParam("The id of the helpdesk ticket you want to add labels to."),
			"labels": toolset.makeArrayParam("The labels to add to the helpdesk ticket. Make sure to only use highly relevant and specific labels."),
		},
	)

	toolset.toolMap["search_slack_messages"] = toolset.makeSchema(
		"search_slack",
		"Searching Slack can be very useful, given that it contains a lot of historical and current communication regarding the projects of the organization.",
		map[string]interface{}{
			"queries": toolset.makeArrayParam("Batch multiple queries for efficient searching. Besides keywords, you can also use Slack's search syntax to refine your search."),
		},
	)

	toolset.toolMap["send_slack_message"] = toolset.makeSchema(
		"send_slack_message",
		"Send or reply to messages via Slack. You can send a message to a user or channel. Useful for when you need to communicate with the outside world.",
		map[string]interface{}{
			"id":      toolset.makeStringParam("The user or channel id to send the message to."),
			"message": toolset.makeStringParam("The content of the direct message you want to send."),
		},
	)

	toolset.toolMap["search_github_code"] = toolset.makeSchema(
		"search_github_code",
		"Search for code on GitHub and retrieve the content of the found files.",
		map[string]interface{}{
			"query": toolset.makeStringParam("The search query to use."),
		},
	)

	toolset.toolMap["add_code_review"] = toolset.makeSchema(
		"add_code_review",
		"Add a code review. Useful for adding a code review to the github system.",
		map[string]interface{}{
			"pr":       toolset.makeStringParam("The pr to review."),
			"review":   toolset.makeStringParam("The review of the pr."),
			"approved": toolset.makeBoolParam("Whether the pr is approved or not."),
		},
	)

	toolset.toolMap["use_browser"] = toolset.makeSchema(
		"use_browser",
		"Use the browser to navigate the web and execute JavaScript to interact with the page. The JavaScript function should return a string containing the desired output.",
		map[string]interface{}{
			"url":        toolset.makeStringParam("The URL to navigate to."),
			"javascript": toolset.makeStringParam("A JavaScript function that returns a string. This function will be executed on the page via the developer console. Example: '() => { return document.title; }' to get the page title. Be mindful of context length, and if you want to retrieve page content, which can be very large, you should make sure your script includes a way to strip away useless content."),
		},
	)

	toolset.toolMap["start_environment"] = toolset.makeSchema(
		"start_environment",
		"Start an new Debian Linux environment. Useful for when you need a full sandbox for writing code or using a terminal.",
		map[string]interface{}{
			"name":   toolset.makeStringParam("The name of the environment to start."),
			"reason": toolset.makeStringParam("The reason you need the environment."),
		},
	)
}

func (toolset *Toolset) makeBoolParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

func (toolset *Toolset) makeStringParam(description string) map[string]string {
	return map[string]string{
		"type":        "string",
		"description": description,
	}
}

func (toolset *Toolset) makeIntParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
}

func (toolset *Toolset) makeEnumParam(description string, values []string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"enum":        values,
		"description": description,
	}
}

func (toolset *Toolset) makeArrayParam(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items":       toolset.makeStringParam("The items in the array."),
	}
}

func (toolset *Toolset) makeSchema(name, description string, params map[string]interface{}) openai.ChatCompletionToolParam {
	return ai.MakeTool(
		name,
		description,
		openai.FunctionParameters{
			"type":       "object",
			"properties": params,
			"required":   toolset.getKeys(params),
		},
	)
}

func (toolset *Toolset) getKeys(params map[string]interface{}) []string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	return keys
}
