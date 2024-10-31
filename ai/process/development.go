package process

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
)

type Development struct {
	CurrentTask    Task     `json:"current_task"`
	Implementation Fragment `json:"implementation"`
	TerminalState  string   `json:"terminal_state"`
}

func NewDevelopment() *Development {
	return &Development{}
}

func (dev *Development) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.development.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", dev.GenerateSchema())
	return prompt
}

func (dev *Development) GenerateSchema() string {
	schema := jsonschema.Reflect(&Development{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(out)
}
