package tools

import (
	"context"
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/errnie"
)

type Recruit struct {
	ToolName     string   `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool to use,enum=recruit,required"`
	Role         string   `json:"role" jsonschema:"title=Role,description=The role to recruit for,required"`
	SystemPrompt string   `json:"system_prompt" jsonschema:"title=System Prompt,description=The system prompt to use,required"`
	Toolset      []string `json:"toolset" jsonschema:"title=Toolset,description=The toolset to use,required,type=array,items=string,uniqueItems=true,anyOf=[environment,helpdesk,boards,browser,recruit,github,neo4j,qdrant,slack,wiki,none]"`
}

func NewRecruit() *Recruit {
	return &Recruit{}
}

func (recruit *Recruit) Use(ctx context.Context, args map[string]any) string {
	return "new agent recruited"
}

func (recruit *Recruit) GenerateSchema() string {
	schema := jsonschema.Reflect(&Recruit{})
	return string(errnie.SafeMust(func() ([]byte, error) {
		return json.MarshalIndent(schema, "", "  ")
	}))
}
