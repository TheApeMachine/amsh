package tools

import (
	"context"
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/errnie"
)

type Inspect struct {
	ToolName string `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool to use,enum=inspect,required"`
	Scope    string `json:"scope" jsonschema:"title=Scope,description=The scope of the inspection,enum=team,required"`
}

func (i *Inspect) Use(ctx context.Context, args map[string]any) string {
	return "inspect"
}

func NewInspect() *Inspect {
	return &Inspect{}
}

func (inspect *Inspect) GenerateSchema() string {
	schema := jsonschema.Reflect(&Inspect{})
	return string(errnie.SafeMust(func() ([]byte, error) {
		return json.MarshalIndent(schema, "", "  ")
	}))
}
