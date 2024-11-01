package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
GenerateSchema is a generic function that generates the JSON schema for
an object that has jsonschema struct tags
*/
func GenerateSchema[T any]() string {
	var instance T
	return string(errnie.SafeMust(func() ([]byte, error) {
		return json.MarshalIndent(jsonschema.Reflect(&instance), "", "  ")
	}))
}

func SystemPrompt(key, schema string) string {
	return JoinWith("\n",
		viper.GetString(fmt.Sprintf("ai.setups.%s.processes.task_analysis.prompt", key)),
		viper.GetString(fmt.Sprintf("ai.setups.%s.personas.task_analyzer.prompt", key)),
		strings.ReplaceAll(
			viper.GetString(fmt.Sprintf("ai.setups.%s.templates.schemas", key)),
			"{{schemas}}",
			schema,
		),
	)
}
