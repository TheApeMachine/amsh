package mastercomputer

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
	"github.com/theapemachine/amsh/errnie"
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
				"unsubscribe_channel",
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

/*
Use a tool, based on the tool call passed in.
*/
func (toolset *Toolset) Use(toolCall openai.ChatCompletionMessageToolCall) (out openai.ChatCompletionToolMessageParam, err error) {
	args := map[string]any{}

	if err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return out, errnie.Error(err)
	}

	switch toolCall.Function.Name {
	case "get_topics":
		topics := twoface.NewQueue().GetTopics()
		return openai.ToolMessage(toolCall.ID, "[TOPICS]\n"+strings.Join(topics, "\n")+"\n[/TOPICS]\n"), nil
	case "topic":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "unsubscribe_channel":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "send_message":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "broadcast_message":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "publish_message":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "worker":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "add_vector_memory":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "search_vector_memory":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "delete_vector_memory":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "add_graph_memory":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "search_graph_memory":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "delete_graph_memory":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "search_work_items":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "get_work_item":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "manage_work_item":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "search_helpdesk_tickets":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "get_helpdesk_labels":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "add_helpdesk_labels":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "search_slack_messages":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "send_slack_channel_message":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "send_slack_user_message":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "search_github_code":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "add_code_review":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
	case "use_browser":
		return openai.ToolMessage(toolCall.ID, "tool not implemented"), nil
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

	toolset.toolMap["unsubscribe_channel"] = toolset.makeSchema(
		"unsubscribe_channel",
		"Unsubscribe from a topic channel.",
		map[string]interface{}{
			"channel": toolset.makeStringParam("The topic channel to unsubscribe from."),
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
		"Search the helpdesk tickets. Useful for retrieving helpdesk tickets from the helpdesk system.",
		map[string]interface{}{
			"query": toolset.makeStringParam("The query to search the helpdesk tickets for."),
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
		"Search the github code. Useful for retrieving github code from the github system.",
		map[string]interface{}{
			"query": toolset.makeStringParam("The query to search the github code for."),
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
