package mastercomputer

import (
	"encoding/json"
	"errors"

	"github.com/openai/openai-go"
	"github.com/theapemachine/amsh/integration/boards"
	"github.com/theapemachine/amsh/integration/trengo"
)

type Tool interface {
	Call(args map[string]any) (string, error)
	Schema() openai.ChatCompletionToolParam
}

// Toolset represents a set of tools available to a worker.
type Toolset struct {
	tools []openai.ChatCompletionToolParam
}

// toolsMap is a map of tool names to tool definitions.
var toolsMap = map[string]Tool{
	"publish_message": &Messaging{},
	"worker":          &Worker{},
}

func init() {
	// Worker
	toolsMap["worker"] = &Worker{}
	toolsMap["messaging"] = &Messaging{}

	// Boards
	toolsMap["search_wiki"] = &boards.SearchWiki{}
	toolsMap["create_workitem"] = &boards.CreateWorkitemSrv{}
	toolsMap["search_workitems"] = &boards.SearchWorkitemsSrv{}
	toolsMap["get_workitem"] = &boards.GetWorkItemSrv{}

	// Trengo
	toolsMap["list_labels"] = &trengo.ListLabels{}
	toolsMap["assign_label"] = &trengo.AssignLabels{}
}

// NewToolset returns a new Toolset for the given key.
func NewToolset(key string) *Toolset {
	toolsets := map[string][]string{
		"reasoning": {
			"publish_message", "worker",
		},
		"messaging": {
			"publish_message",
		},
		"boards": {
			"create_workitem", "search_workitems", "get_workitem", "search_wiki",
		},
		"trengo": {
			"list_labels", "assign_label",
		},
	}

	if tools, exists := toolsets[key]; exists {
		var selectedTools []openai.ChatCompletionToolParam
		for _, toolKey := range tools {
			if tool, ok := toolsMap[toolKey]; ok {
				selectedTools = append(selectedTools, tool.Schema())
			}
		}
		return &Toolset{tools: selectedTools}
	}

	return nil
}

/*
Use a tool, based on the tool call passed in.
*/
func UseTool(toolCall openai.ChatCompletionMessageToolCall) (string, error) {
	tool, ok := toolsMap[toolCall.Function.Name]
	if !ok {
		return "", errors.New("tool not found")
	}

	args := map[string]any{}

	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", err
	}

	return tool.Call(args)
}
