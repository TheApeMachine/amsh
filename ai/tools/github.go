package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/integration/git"
	"github.com/theapemachine/errnie"
)

type Github struct {
	ToolName  string `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool to use,enum=github,required"`
	Operation string `json:"operation" jsonschema:"title=Operation,description=The operation to perform,enum=search_code,required"`
	Query     string `json:"query" jsonschema:"title=Query,description=The query to search for,required"`
	hub       *git.Hub
}

func NewGithub() *Github {
	return &Github{
		hub: git.NewHub(),
	}
}

func (github *Github) GenerateSchema() string {
	schema := jsonschema.Reflect(&Github{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (github *Github) Use(ctx context.Context, args map[string]any) string {
	switch github.Operation {
	case "search_code":
		results := errnie.SafeMust(func() ([]git.CodeSearchResult, error) {
			return github.hub.SearchCode(ctx, github.Query)
		})

		return fmt.Sprintf("%v", results)
	}

	return ""
}
