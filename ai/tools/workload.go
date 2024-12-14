package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/errnie"
)

/*
Workload is a tool that creates a new workload, which is a JSON schema that describes a workload.
*/
type Workload struct {
	ToolName   string `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool to use,enum=workload,required"`
	Name       string `json:"name" jsonschema:"title=Name,description=The name of the new workload,required"`
	JSONSchema string `json:"json_schema" jsonschema:"title=JSON Schema,description=The JSON schema to use for the workload,required"`
}

/*
Use the Process tool to create a new process, which will write a JSON schema to the file system.
*/
func (workload *Workload) Use(ctx context.Context, args map[string]any) string {
	errnie.Log("using workload tool %v", args)
	path := filepath.Join(os.Getenv("HOME"), ".amsh", "workloads")

	// Create the directory if it doesn't exist
	os.MkdirAll(path, 0755)

	// Write the JSON schema to the file
	errnie.MustVoid(os.WriteFile(filepath.Join(path, workload.Name+".json"), []byte(workload.JSONSchema), 0644))

	return "Workload created"
}

func NewWorkload() *Workload {
	return &Workload{}
}

func (workload *Workload) GenerateSchema() string {
	schema := jsonschema.Reflect(&Workload{})
	return string(errnie.SafeMust(func() ([]byte, error) {
		return json.MarshalIndent(schema, "", "  ")
	}))
}
