package trengo

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
	tools.schemas["list_labels"] = tools.GetLabels()
	tools.schemas["assign_label"] = tools.GetLabel()

	return tools.schemas
}

func (t *Tools) GetLabels() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"list_labels",
		"List all the labels in a way that the language model can understand.",
		openai.FunctionParameters{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
	)
}

func (t *Tools) GetLabel() openai.ChatCompletionToolParam {
	return ai.MakeTool(
		"assign_label",
		"Assign a label to a ticket.",
		openai.FunctionParameters{
			"type": "object",
			"properties": map[string]interface{}{
				"label_id": map[string]string{
					"type":        "string",
					"description": "The ID of the label to assign.",
				},
				"ticket_id": map[string]string{
					"type":        "string",
					"description": "The ID of the ticket to assign the label to.",
				},
			},
			"required": []string{"label_id", "ticket_id"},
		},
	)
}
