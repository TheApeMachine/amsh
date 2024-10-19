package mastercomputer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/integration/boards"
	"github.com/theapemachine/amsh/integration/comms"
	"github.com/theapemachine/amsh/integration/git"
	"github.com/theapemachine/amsh/integration/trengo"
	"github.com/theapemachine/amsh/twoface"
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
				"get_topics",
				"topic",
				"send_message",
				"broadcast_message",
				"publish_message",
				"add_vector_memory",
				"search_vector_memory",
				"delete_vector_memory",
				"add_graph_memory",
				"search_graph_memory",
				"delete_graph_memory",
			},
			workloads: map[string][]string{
				"managing": {
					"search_work_items",
					"get_work_item",
					"manage_work_item",
				},
				"researching": {
					"search_github_code",
					"search_slack_messages",
					"search_helpdesk_tickets",
					"search_work_items",
					"get_helpdesk_messages",
					"get_helpdesk_message",
				},
				"executing": {
					"start_environment",
					"search_github_code",
					"add_helpdesk_labels",
					"get_helpdesk_labels",
				},
				"reviewing": {
					"add_code_review",
				},
				"communicating": {
					"search_slack_messages",
					"send_slack_channel_message",
					"send_slack_user_message",
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
	if _, ok := args[arg]; ok {
		return true
	}
	return false
}

/*
Use a tool, based on the tool call passed in.
*/
func (toolset *Toolset) Use(ID string, toolCall openai.ChatCompletionMessageToolCall) (out openai.ChatCompletionToolMessageParam, err error) {
	args := map[string]any{}

	if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return out, errnie.Error(err)
	}

	switch toolCall.Function.Name {
	case "get_topics":
		topics := twoface.NewQueue().GetTopics()
		return openai.ToolMessage(toolCall.ID, "[TOPICS]\n"+strings.Join(topics, "\n")+"\n[/TOPICS]\n"), nil
	case "topic":
		// Check if we are subscribing or unsubscribing
		if toolset.getArgPresent(args, "action") && toolset.getArgPresent(args, "topic_channel") {
			if args["action"] == "subscribe" {
				twoface.NewQueue().Subscribe(ID, args["topic_channel"].(string))
				return openai.ToolMessage(toolCall.ID, "subscribed to "+args["topic_channel"].(string)), nil
			} else if args["action"] == "unsubscribe" {
				twoface.NewQueue().Unsubscribe(ID, args["topic_channel"].(string))
				return openai.ToolMessage(toolCall.ID, "unsubscribed from "+args["topic_channel"].(string)), nil
			}
		}

		return openai.ToolMessage(toolCall.ID, "something went wrong"), nil
	case "send_message":
		if toolset.getArgPresent(args, "to") && toolset.getArgPresent(args, "subject") && toolset.getArgPresent(args, "message") {
			twoface.NewQueue().PubCh <- data.New(ID, args["subject"].(string), args["to"].(string), []byte(args["message"].(string)))
			return openai.ToolMessage(toolCall.ID, "message sent"), nil
		}

		return openai.ToolMessage(toolCall.ID, "something went wrong"), nil
	case "broadcast_message":
		if toolset.getArgPresent(args, "subject") && toolset.getArgPresent(args, "message") {
			twoface.NewQueue().PubCh <- data.New(ID, args["subject"].(string), "broadcast", []byte(args["message"].(string)))
			return openai.ToolMessage(toolCall.ID, "message broadcasted"), nil
		}

		return openai.ToolMessage(toolCall.ID, "something went wrong"), nil
	case "publish_message":
		if toolset.getArgPresent(args, "channel") && toolset.getArgPresent(args, "subject") && toolset.getArgPresent(args, "message") {
			twoface.NewQueue().PubCh <- data.New(ID, args["subject"].(string), args["channel"].(string), []byte(args["message"].(string)))
			return openai.ToolMessage(toolCall.ID, "message published"), nil
		}

		return openai.ToolMessage(toolCall.ID, "something went wrong"), nil
	case "worker":
		if toolset.getArgPresent(args, "system_prompt") && toolset.getArgPresent(args, "user_prompt") && toolset.getArgPresent(args, "toolset") {
			builder := NewBuilder()
			worker := builder.NewWorker(builder.GetRole(args["toolset"].(string)))
			worker.Start()
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("%s (%s) created", worker.buffer.Peek("role"), worker.name)), nil
		}

		return openai.ToolMessage(toolCall.ID, "something went wrong"), nil
	case "add_vector_memory":
		if toolset.getArgPresent(args, "content") && toolset.getArgPresent(args, "private") {
			longTerm := memory.NewLongTerm(ID)
			content := args["content"].(string)
			private := args["private"].(bool)

			// Implement the logic to add vector memory
			// This is a placeholder implementation
			result, err := longTerm.Query("vector", fmt.Sprintf("ADD %s %v", content, private))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error adding vector memory: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Vector memory added: %v", result)), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for add_vector_memory"), nil

	case "search_vector_memory":
		if toolset.getArgPresent(args, "question") && toolset.getArgPresent(args, "private") {
			longTerm := memory.NewLongTerm(ID)
			question := args["question"].(string)
			private := args["private"].(bool)

			// Implement the logic to search vector memory
			result, err := longTerm.Query("vector", fmt.Sprintf("SEARCH %s %v", question, private))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching vector memory: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Vector memory search results: %v", result)), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_vector_memory"), nil

	case "delete_vector_memory":
		if toolset.getArgPresent(args, "id") {
			longTerm := memory.NewLongTerm(ID)
			id := args["id"].(string)

			// Implement the logic to delete vector memory
			result, err := longTerm.Query("vector", fmt.Sprintf("DELETE %s", id))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error deleting vector memory: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Vector memory deleted: %v", result)), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for delete_vector_memory"), nil

	case "add_graph_memory":
		if toolset.getArgPresent(args, "private") && toolset.getArgPresent(args, "cypher") {
			longTerm := memory.NewLongTerm(ID)
			_ = args["private"].(bool)
			cypher := args["cypher"].(string)

			// Implement the logic to add graph memory
			result, err := longTerm.Write("graph", cypher)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error adding graph memory: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Graph memory added: %v", result)), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for add_graph_memory"), nil

	case "search_graph_memory":
		if toolset.getArgPresent(args, "cypher") && toolset.getArgPresent(args, "private") {
			longTerm := memory.NewLongTerm(ID)
			cypher := args["cypher"].(string)
			_ = args["private"].(bool)

			// Implement the logic to search graph memory
			result, err := longTerm.Query("graph", cypher)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching graph memory: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Graph memory search results: %v", result)), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_graph_memory"), nil

	case "delete_graph_memory":
		if toolset.getArgPresent(args, "cypher") {
			longTerm := memory.NewLongTerm(ID)
			cypher := args["cypher"].(string)

			// Implement the logic to delete graph memory
			result, err := longTerm.Query("graph", fmt.Sprintf("DELETE %s", cypher))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error deleting graph memory: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, fmt.Sprintf("Graph memory deleted: %v", result)), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for delete_graph_memory"), nil
	case "search_work_items":
		if toolset.getArgPresent(args, "query") {
			query := args["query"].(string)
			ctx := context.Background()
			searchSrv, err := boards.NewSearchWorkitemsSrv(ctx, "your_project_name")
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error creating search service: "+err.Error()), nil
			}
			result, err := searchSrv.SearchWorkitems(ctx, query)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching work items: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, result), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_work_items"), nil

	case "get_work_item":
		if toolset.getArgPresent(args, "id") {
			id, err := strconv.Atoi(args["id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid work item ID: "+err.Error()), nil
			}
			ctx := context.Background()
			getSrv, err := boards.NewGetWorkItemSrv(ctx, os.Getenv("AZDO_ORG_URL"), os.Getenv("AZDO_PAT"))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error creating get service: "+err.Error()), nil
			}
			result, err := getSrv.GetWorkitem(ctx, id)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error getting work item: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, result), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for get_work_item"), nil

	case "manage_work_item":
		if toolset.getArgPresent(args, "action") && toolset.getArgPresent(args, "title") && toolset.getArgPresent(args, "description") {
			action := args["action"].(string)
			title := args["title"].(string)
			description := args["description"].(string)
			ctx := context.Background()

			switch action {
			case "create":
				createSrv, err := boards.NewCreateWorkitemSrv(ctx, "playground")
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error creating create service: "+err.Error()), nil
				}
				result, err := createSrv.CreateWorkitem(ctx, title, description)
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error creating work item: "+err.Error()), nil
				}
				return openai.ToolMessage(toolCall.ID, result), nil
			case "update":
				if !toolset.getArgPresent(args, "id") {
					return openai.ToolMessage(toolCall.ID, "ID is required for updating a work item"), nil
				}
				id, err := strconv.Atoi(args["id"].(string))
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Invalid work item ID: "+err.Error()), nil
				}
				updateSrv, err := boards.NewUpdateWorkitemSrv(ctx, "playground")
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error creating update service: "+err.Error()), nil
				}
				result, err := updateSrv.UpdateWorkitem(ctx, id, title, description)
				if err != nil {
					return openai.ToolMessage(toolCall.ID, "Error updating work item: "+err.Error()), nil
				}
				return openai.ToolMessage(toolCall.ID, result), nil
			default:
				return openai.ToolMessage(toolCall.ID, "Invalid action for manage_work_item"), nil
			}
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for manage_work_item"), nil

	case "search_helpdesk_tickets":
		if toolset.getArgPresent(args, "query") && toolset.getArgPresent(args, "page") {
			query := args["query"].(string)
			page, err := strconv.Atoi(args["page"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid page number: "+err.Error()), nil
			}

			ticketService := trengo.NewTicketService()
			tickets, err := ticketService.ListTickets(context.Background(), page)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching helpdesk tickets: "+err.Error()), nil
			}

			// Format the results for the AI
			formattedResults := fmt.Sprintf("Helpdesk Tickets (Page %d, Query: %s):\n\n", page, query)
			for _, ticket := range tickets {
				formattedResults += fmt.Sprintf("ID: %d\nSubject: %s\nStatus: %s\nCreated At: %s\nUpdated At: %s\nAssigned To: %s\nLast Message: %s\n\n",
					ticket.ID, ticket.Subject, ticket.Status, ticket.CreatedAt, ticket.UpdatedAt, ticket.AssignedTo, ticket.LastMessage)
			}

			return openai.ToolMessage(toolCall.ID, formattedResults), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_helpdesk_tickets"), nil

	case "get_helpdesk_messages":
		if toolset.getArgPresent(args, "ticket_id") {
			ticketID, err := strconv.Atoi(args["ticket_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid ticket ID: "+err.Error()), nil
			}

			messageService := trengo.NewMessageService(os.Getenv("TRENGO_BASE_URL"), os.Getenv("TRENGO_API_TOKEN"))
			messages, err := messageService.ListMessages(context.Background(), ticketID)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error getting helpdesk messages: "+err.Error()), nil
			}

			// Format the results for the AI
			formattedResults := fmt.Sprintf("Helpdesk Messages for Ticket %d:\n\n", ticketID)
			for _, message := range messages {
				formattedResults += fmt.Sprintf("ID: %d\nBody: %s\nCreated At: %s\nUser: %s (ID: %d)\n\n",
					message.ID, message.Body, message.CreatedAt, message.User.Name, message.User.ID)
			}

			return openai.ToolMessage(toolCall.ID, formattedResults), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for get_helpdesk_messages"), nil

	case "get_helpdesk_message":
		if toolset.getArgPresent(args, "ticket_id") && toolset.getArgPresent(args, "message_id") {
			ticketID, err := strconv.Atoi(args["ticket_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid ticket ID: "+err.Error()), nil
			}
			messageID, err := strconv.Atoi(args["message_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid message ID: "+err.Error()), nil
			}

			messageService := trengo.NewMessageService(os.Getenv("TRENGO_BASE_URL"), os.Getenv("TRENGO_API_TOKEN"))
			message, err := messageService.FetchMessage(context.Background(), ticketID, messageID)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error getting helpdesk message: "+err.Error()), nil
			}

			// Format the result for the AI
			formattedResult := fmt.Sprintf("Helpdesk Message (Ticket %d, Message %d):\n\nBody: %s\nCreated At: %s\nUser: %s (ID: %d)",
				ticketID, messageID, message.Body, message.CreatedAt, message.User.Name, message.User.ID)

			return openai.ToolMessage(toolCall.ID, formattedResult), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for get_helpdesk_message"), nil

	case "get_helpdesk_labels":
		ctx := context.Background()
		listLabels := trengo.NewListLabels()
		labels, err := listLabels.List(ctx)
		if err != nil {
			return openai.ToolMessage(toolCall.ID, "Error getting helpdesk labels: "+err.Error()), nil
		}
		labelsJSON, err := json.Marshal(labels)
		if err != nil {
			return openai.ToolMessage(toolCall.ID, "Error marshaling labels: "+err.Error()), nil
		}
		return openai.ToolMessage(toolCall.ID, string(labelsJSON)), nil
	case "add_helpdesk_labels":
		if toolset.getArgPresent(args, "ticket_id") && toolset.getArgPresent(args, "label_id") {
			ticketID, err := strconv.Atoi(args["ticket_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid ticket ID: "+err.Error()), nil
			}
			labelID, err := strconv.Atoi(args["label_id"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Invalid label ID: "+err.Error()), nil
			}
			ctx := context.Background()
			assignLabels := trengo.NewAssignLabels()
			err = assignLabels.Attach(ctx, labelID, ticketID)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error adding helpdesk label: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, "Label successfully added to the ticket"), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for add_helpdesk_labels"), nil
	case "search_slack_messages":
		if toolset.getArgPresent(args, "query") {
			search := comms.NewSearch()
			messages, err := search.SearchMessages(context.Background(), args["query"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching Slack messages: "+err.Error()), nil
			}
			result, _ := json.Marshal(messages)
			return openai.ToolMessage(toolCall.ID, string(result)), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_slack_messages"), nil
	case "send_slack_channel_message":
		if toolset.getArgPresent(args, "channel") && toolset.getArgPresent(args, "message") {
			msg := comms.NewMessage(args["channel"].(string), args["message"].(string))
			err := msg.Post(context.Background())
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error sending Slack channel message: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, "Slack channel message sent successfully"), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for send_slack_channel_message"), nil
	case "send_slack_user_message":
		if toolset.getArgPresent(args, "user") && toolset.getArgPresent(args, "message") {
			// Note: This assumes the user ID is provided. You might need to look up the user ID from the username if that's what's provided.
			msg := comms.NewMessage(args["user"].(string), args["message"].(string))
			err := msg.Post(context.Background())
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error sending Slack user message: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, "Slack user message sent successfully"), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for send_slack_user_message"), nil
	case "search_github_code":
		if toolset.getArgPresent(args, "query") {
			hub := git.NewHub()
			results, err := hub.SearchCode(context.Background(), args["query"].(string))
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error searching GitHub code: "+err.Error()), nil
			}

			// Format the results for the AI
			formattedResults := "GitHub Code Search Results:\n\n"
			for _, result := range results {
				formattedResults += fmt.Sprintf("Repository: %s\nFile: %s\nContent:\n```\n%s\n```\n\n",
					result.Repository, result.Path, result.Content)
			}

			return openai.ToolMessage(toolCall.ID, formattedResults), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for search_github_code"), nil
	case "add_code_review":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "use_browser":
		if toolset.getArgPresent(args, "url") && toolset.getArgPresent(args, "javascript") {
			browser := NewBrowser()
			result, err := browser.Run(args)
			if err != nil {
				return openai.ToolMessage(toolCall.ID, "Error using browser: "+err.Error()), nil
			}
			return openai.ToolMessage(toolCall.ID, result), nil
		}
		return openai.ToolMessage(toolCall.ID, "Invalid arguments for use_browser"), nil
	case "start_environment":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	}

	return openai.ChatCompletionToolMessageParam{}, err
}

func (toolset *Toolset) makeTools() {
	toolset.toolMap["get_topics"] = toolset.makeSchema(
		"get_topics",
		"Get a list of all topic channels currently in use.",
		map[string]interface{}{},
	)

	toolset.toolMap["topic"] = toolset.makeSchema(
		"topic",
		"Manage your subscriptions to topic channels.",
		map[string]interface{}{
			"topic_channel": toolset.makeStringParam("The topic channel to subscribe to."),
			"action":        toolset.makeEnumParam("The action to perform.", []string{"subscribe", "unsubscribe"}),
		},
	)

	toolset.toolMap["send_message"] = toolset.makeSchema(
		"send_message",
		"Send a message to a user or a channel.",
		map[string]interface{}{
			"to":       toolset.makeStringParam("The username of the user or channel to send the message to."),
			"subject":  toolset.makeStringParam("The subject of the message you want to send."),
			"priority": toolset.makeEnumParam("The priority of the message you want to send.", []string{"low", "normal", "high"}),
			"message":  toolset.makeStringParam("The content of the message you want to send."),
		},
	)

	toolset.toolMap["broadcast_message"] = toolset.makeSchema(
		"broadcast_message",
		"Broadcast a message to all users.",
		map[string]interface{}{
			"subject":  toolset.makeStringParam("The subject of the message you want to send."),
			"message":  toolset.makeStringParam("The content of the message you want to send."),
			"priority": toolset.makeEnumParam("The priority of the message you want to send.", []string{"low", "normal", "high"}),
		},
	)

	toolset.toolMap["publish_message"] = toolset.makeSchema(
		"publish_message",
		"Publish a message to a topic channel.",
		map[string]interface{}{
			"channel":  toolset.makeStringParam("The channel to publish the message to."),
			"subject":  toolset.makeStringParam("The subject of the message you want to publish."),
			"message":  toolset.makeStringParam("The content of the message you want to publish."),
			"priority": toolset.makeEnumParam("The priority of the message you want to publish.", []string{"low", "normal", "high"}),
		},
	)

	toolset.toolMap["worker"] = toolset.makeSchema(
		"worker",
		"Create any type of worker, and delegate sub-tasks to them.",
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
			"private": toolset.makeBoolParam("Whether this is a private memory or not."),
		},
	)

	toolset.toolMap["search_vector_memory"] = toolset.makeSchema(
		"search_vector_memory",
		"Search the vector store. Useful for retrieving memories from the vector store.",
		map[string]interface{}{
			"question": toolset.makeStringParam("The question to search the vector memory for."),
			"private":  toolset.makeBoolParam("Whether to search private memories or not."),
		},
	)

	toolset.toolMap["delete_vector_memory"] = toolset.makeSchema(
		"delete_vector_memory",
		"Delete a memory from the vector store. Useful for deleting private memories from the vector store.",
		map[string]interface{}{
			"id": toolset.makeStringParam("The id of the memory to delete from the vector store."),
		},
	)

	toolset.toolMap["add_graph_memory"] = toolset.makeSchema(
		"add_graph_memory",
		"Add a memory to the graph store. Useful for storing memories as connected relationships.",
		map[string]interface{}{
			"private": toolset.makeBoolParam("Whether this is a private memory or not."),
			"cypher":  toolset.makeStringParam("The cypher query to add to the graph store."),
		},
	)

	toolset.toolMap["search_graph_memory"] = toolset.makeSchema(
		"search_graph_memory",
		"Search the graph store. Useful for retrieving memories from the graph store.",
		map[string]interface{}{
			"cypher":  toolset.makeStringParam("The cypher query to search the graph memory for."),
			"private": toolset.makeBoolParam("Whether to search private memories or not."),
		},
	)

	toolset.toolMap["delete_graph_memory"] = toolset.makeSchema(
		"delete_graph_memory",
		"Delete a memory from the graph store. Useful for deleting private memories from the graph store.",
		map[string]interface{}{
			"cypher": toolset.makeStringParam("The cypher query to delete from the graph store."),
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
		"Manage a work item. Useful for creating, updating, and deleting work items in the project management system.",
		map[string]interface{}{
			"action":      toolset.makeEnumParam("The action to perform on the work item.", []string{"create", "update"}),
			"title":       toolset.makeStringParam("The title of the work item."),
			"description": toolset.makeStringParam("The description of the work item."),
			"priority":    toolset.makeEnumParam("The priority of the work item.", []string{"low", "normal", "high"}),
			"tags":        toolset.makeArrayParam("The tags of the work item."),
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

	toolset.toolMap["get_helpdesk_message"] = toolset.makeSchema(
		"get_helpdesk_message",
		"Get a specific message from a helpdesk ticket in Trengo.",
		map[string]interface{}{
			"ticket_id":  toolset.makeIntParam("The ID of the ticket the message belongs to."),
			"message_id": toolset.makeIntParam("The ID of the message to fetch."),
		},
	)

	toolset.toolMap["get_helpdesk_labels"] = toolset.makeSchema(
		"get_helpdesk_labels",
		"Get the available labels for a helpdesk ticket. Useful for knowing which labels exist.",
		map[string]interface{}{
			"id": toolset.makeStringParam("The id of the helpdesk ticket you want to get labels for."),
		},
	)

	toolset.toolMap["add_helpdesk_labels"] = toolset.makeSchema(
		"add_helpdesk_labels",
		"Add labels to a helpdesk ticket. Useful for categorizing helpdesk tickets.",
		map[string]interface{}{
			"id":     toolset.makeStringParam("The id of the helpdesk ticket you want to add labels to."),
			"labels": toolset.makeArrayParam("The labels to add to the helpdesk ticket."),
		},
	)

	toolset.toolMap["search_slack_messages"] = toolset.makeSchema(
		"search_slack_messages",
		"Search the slack messages. Useful for retrieving slack messages from the slack system.",
		map[string]interface{}{
			"query": toolset.makeStringParam("The query to search the slack messages for."),
		},
	)

	toolset.toolMap["send_slack_channel_message"] = toolset.makeSchema(
		"send_slack_channel_message",
		"Send a message to a channel in slack.",
		map[string]interface{}{
			"channel": toolset.makeStringParam("The channel to send the message to."),
			"message": toolset.makeStringParam("The content of the message you want to send."),
		},
	)

	toolset.toolMap["send_slack_user_message"] = toolset.makeSchema(
		"send_slack_user_message",
		"Send a direct message to a user in slack.",
		map[string]interface{}{
			"user":    toolset.makeStringParam("The user to send the direct message to."),
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
		"Use the browser to navigate the web. Useful for when you need to navigate the web.",
		map[string]interface{}{
			"url":        toolset.makeStringParam("The url to navigate to."),
			"javascript": toolset.makeStringParam("The JavaScript to run on the page via the developer console. Must be a valid javascript function that returns a string."),
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
