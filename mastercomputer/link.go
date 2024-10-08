package mastercomputer

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/utils"
)

/*
Link connects agents to each other in various ways.
Agents can be receivers or senders only, or both.
*/
type Link struct {
	ID        string
	Senders   []string
	Receivers []string
	Function  *openai.FunctionDefinition
}

/*
NewLink creates a new link.
*/
func NewLink() *Link {
	return &Link{
		ID:        utils.NewID(),
		Senders:   make([]string, 0),
		Receivers: make([]string, 0),
		Function: &openai.FunctionDefinition{
			Name:        "link",
			Description: "Use to connect things together, which opens up communication channels, for enhanced collaboration.",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Description:          "Use to connect things together, which opens up communication channels, for enhanced collaboration.",
				Properties: map[string]jsonschema.Definition{
					"senders": {
						Type:        jsonschema.Array,
						Description: "The list of senders. depending on your requirements, any linked object should be either a sender, receiver, or both.",
						Items: &jsonschema.Definition{
							Type: jsonschema.String,
						},
					},
					"receivers": {
						Type:        jsonschema.Array,
						Description: "The list of receivers, depending on your requirements, any linked object should be either a sender, receiver, or both.",
						Items: &jsonschema.Definition{
							Type: jsonschema.String,
						},
					},
				},
				Required: []string{"senders", "receivers"},
			},
		},
	}
}
