package process

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Discussion struct {
	NextSpeaker string `json:"next_speaker" jsonschema:"title=Next Speaker,description=The name of the next speaker; required"`
}

func NewDiscussion() *Discussion {
	return &Discussion{}
}

func (discussion *Discussion) GenerateSchema() string {
	schema := jsonschema.Reflect(&Discussion{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}
