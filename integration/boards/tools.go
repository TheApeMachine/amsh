package boards

import (
	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/ai"
)

type Tools struct {
	schemas map[string]openai.ChatCompletionToolParam
}

func NewTools() *Tools {
	return &Tools{
		schemas: make(map[string]openai.ChatCompletionToolParam),
	}
}

func (tools *Tools) GetSchemas() map[string]openai.ChatCompletionToolParam {
	tools.schemas["search_tickets"] = tools.SearchTickets()
	tools.schemas["get_ticket"] = tools.GetTicket()
	tools.schemas["create_ticket"] = tools.CreateTicket()

	return tools.schemas
}

func (t *Tools) SearchTickets() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"search_tickets",
		"Search for tickets based on a set of criteria.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]string{
					"type":        "string",
					"description": "The search query to use.",
				},
			},
			"required": []string{"query"},
		},
	)
}

func (t *Tools) GetTicket() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"get_ticket",
		"Get a ticket by ID.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"ticket_id": map[string]string{
					"type":        "string",
					"description": "The ID of the ticket to get.",
				},
				"fields": map[string]string{
					"type":        "string",
					"description": "The fields to get.",
				},
			},
			"required": []string{"ticket_id", "fields"},
		},
	)
}

func (t *Tools) CreateTicket() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"create_ticket",
		"Create a new ticket.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]string{
					"type":        "string",
					"description": "The title of the ticket.",
				},
				"description": map[string]string{
					"type":        "string",
					"description": "The description of the ticket.",
				},
				"priority": map[string]interface{}{
					"type":        "string",
					"description": "The priority of the ticket.",
					"enum":        []string{"1", "2", "3", "4"},
				},
			},
			"required": []string{"title", "description", "priority"},
		},
	)
}
