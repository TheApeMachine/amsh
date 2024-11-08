package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

/*
Process is a tool that creates a new process, which is a JSON schema that describes a process.
*/
type Process struct {
	ToolName   string `json:"tool_name" jsonschema:"title=Tool Name,description=The name of the tool to use,enum=process,required"`
	Name       string `json:"name" jsonschema:"title=Name,description=The name of the new process,required"`
	JSONSchema string `json:"json_schema" jsonschema:"title=JSON Schema,description=The JSON schema to use for the process,required"`
}

/*
Use the Process tool to create a new process, which will write a JSON schema to the file system.
*/
func (process *Process) Use(ctx context.Context, args map[string]any) string {
	errnie.Log("using process tool %v", args)
	path := filepath.Join("~/.amsh", "processes")

	// Create the directory if it doesn't exist
	os.MkdirAll(path, 0755)

	// Write the JSON schema to the file
	errnie.MustVoid(os.WriteFile(filepath.Join(path, process.Name+".json"), []byte(process.JSONSchema), 0644))

	return "Process created"
}

func NewProcess() *Process {
	return &Process{}
}

func (process *Process) GenerateSchema() string {
	schema := jsonschema.Reflect(&Process{})
	return string(errnie.SafeMust(func() ([]byte, error) {
		return json.MarshalIndent(schema, "", "  ")
	}))
}
