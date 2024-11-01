package process

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
Discussion is a process where multiple AI agents discuss a topic and come to a
consensus, which will be the final response of the process.
*/
type Discussion struct {
	Topic       string `json:"topic" jsonschema:"title=Topic,description=The topic to discuss,required"`
	NextSpeaker string `json:"next_speaker" jsonschema:"title=Next Speaker,description=The name of the next speaker,required"`
}

/*
NewDiscussion returns a Discussion process that can be used to discuss a topic.
*/
func NewDiscussion() *Discussion {
	return &Discussion{}
}

/*
SystemPrompt returns the system prompt for the Discussion process.
*/
func (discussion *Discussion) SystemPrompt(key string) string {
	log.Info("SystemPrompt", "key", key)
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.trengo.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", discussion.GenerateSchema())
	return prompt
}

/*
GenerateSchema is used as the value to inject into the system prompt, which guides
the AI on how to format its response.
*/
func (discussion *Discussion) GenerateSchema() string {
	var (
		out []byte
		err error
	)

	schema := jsonschema.Reflect(&Discussion{})
	
	if out, err = json.MarshalIndent(schema, "", "  "); err != nil {
		errnie.Error(err)
	}

	return string(out)
}
