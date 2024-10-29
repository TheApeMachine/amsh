package process

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Request struct {
	Resource    string   `json:"resource" jsonschema:"required; description:The resource you want to access; enum:slack,boards,github"`
	Operation   string   `json:"operation" jsonschema:"required; description:The operation to perform on the resource; enum:search"`
	Description string   `json:"description" jsonschema:"required; description:The description of the search you want to perform"`
	Keywords    []string `json:"keywords" jsonschema:"required; description:The keywords to use for the search"`
	Exclusions  []string `json:"exclusions" jsonschema:"required; description:The keywords you want to exclude from the search"`
}

func NewRequest() *Request {
	return &Request{}
}

func (request *Request) GenerateSchema() string {
	schema := jsonschema.Reflect(&Request{})
	out, err := json.MarshalIndent(schema, "", "  ")

	if err != nil {
		errnie.Error(err)
	}

	return string(out)
}
