package process

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/errnie"
)

type Planning struct {
	Epics []Epic `json:"epics" jsonschema:"description:The epics detected in the user's message"`
}

type Epic struct {
	Title       string   `json:"title" jsonschema:"required,description:The title of the epic"`
	Description string   `json:"description" jsonschema:"required,description:The Gherkin description of the epic"`
	Tags        []Tag    `json:"tags" jsonschema:"required,description:The tags to use for the epic"`
	Stories     []Story  `json:"stories" jsonschema:"required,description:The stories that are part of the epic"`
	Assignee    Assignee `json:"assignee" jsonschema:"required,description:The assignee of the epic"`
	Links       []Link   `json:"links" jsonschema:"required,description:The links to use for the epic"`
}

type Story struct {
	Title       string   `json:"title" jsonschema:"required,description:The title of the story"`
	Description string   `json:"description" jsonschema:"required,description:The Ghekrin description of the story"`
	Tags        []Tag    `json:"tags" jsonschema:"required,description:The tags to use for the story"`
	Tasks       []Task   `json:"tasks" jsonschema:"required,description:The tasks that are part of the story"`
	Assignee    Assignee `json:"assignee" jsonschema:"required,description:The assignee of the story"`
}

type Task struct {
	Title       string   `json:"title" jsonschema:"required,description:The title of the task"`
	Description string   `json:"description" jsonschema:"required,description:The Gherkin description of the task"`
	Tags        []Tag    `json:"tags" jsonschema:"required,description:The tags to use for the task"`
	Assignee    Assignee `json:"assignee" jsonschema:"required,description:The assignee of the task"`
}

type Tag struct {
	Tag string `json:"tag" jsonschema:"required,description:The tag to use for the epic or story"`
}

type Assignee struct {
	Username string `json:"username" jsonschema:"required,description:The username of the assignee of the task"`
}

type Link struct {
	FromID string `json:"from_id" jsonschema:"required,description:The ID of the link to use for the epic, story or task"`
	ToID   string `json:"to_id" jsonschema:"required,description:The ID of the link to use for the epic, story or task"`
	Type   string `json:"type" jsonschema:"required,description:The type of the link to use for the epic, story or task"`
}

func NewPlanning() *Planning {
	return &Planning{}
}

/*
Extract the planning JSON from the response.
*/
func (planning *Planning) Extract(response string) *Planning {
	jsonStart := strings.Index(response, "```json")
	jsonEnd := strings.LastIndex(response, "```")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil
	}

	json := response[jsonStart+7 : jsonEnd]

	return planning.Unmarshal([]byte(json))
}

func (planning *Planning) Marshal() string {
	buf, err := json.Marshal(planning)

	if err != nil {
		errnie.Error(err)
	}

	return string(buf)
}

func (planning *Planning) Unmarshal(buf []byte) *Planning {
	err := json.Unmarshal(buf, planning)

	if err != nil {
		errnie.Error(err)
	}

	return planning
}

func (planning *Planning) SystemPrompt(key string) string {
	log.Info("SystemPrompt", "key", key)

	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.slack.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", planning.GenerateSchema())

	request := NewRequest()
	prompt = strings.ReplaceAll(prompt, "{{requests}}", request.GenerateSchema())

	return prompt
}

func (planning *Planning) GenerateSchema() string {
	schema := jsonschema.Reflect(&Planning{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}

	return string(out)
}

/*
Format the process as a pretty-printed JSON string.
*/
func (planning *Planning) Format() string {
	pretty, _ := json.MarshalIndent(planning, "", "  ")
	return string(pretty)
}

/*
String returns a human-readable string representation of the process.
*/
func (planning *Planning) String() {
	berrt.Info("Planning", planning)
}
