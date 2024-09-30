package tweaker

import (
	"github.com/spf13/viper"
)

var cfg *Config

func init() {
	cfg = NewConfig()
}

type Config struct {
	v *viper.Viper
}

func NewConfig() *Config {
	return &Config{
		v: viper.GetViper(),
	}
}

/* LogLevel returns the log level from the config. */
func LogLevel() string { return cfg.v.GetString("loglevel") }

/* Setup returns the setup from the config. */
func Setup() string { return cfg.v.GetString("ai.setup") }

/* PromptTemplate returns the prompt template from the config. */
func PromptTemplate(key string) string { return cfg.v.GetString("ai.prompt.template." + key) }

/* PromptPrefix returns the prompt prefix from the config. */
func PromptPrefix() string { return cfg.v.GetString("ai.prompt.prefix") }

/* PromptSuffix returns the prompt suffix from the config. */
func PromptSuffix() string { return cfg.v.GetString("ai.prompt.suffix") }

/* PromptInstructions returns the prompt instructions from the config. */
func PromptInstructions() string { return cfg.v.GetString("ai.prompt.instructions") }

/* Setups returns the setups from the config. */
func Setups() map[string]interface{} { return cfg.v.GetStringMap("ai.setups." + Setup()) }

/* Agents returns the agents from the config. */
func Agents() []map[string]interface{} {
	agents := cfg.v.Get("ai.setups." + Setup() + ".agents")
	if agents == nil {
		return nil
	}

	// Agents is expected to be a slice of maps.
	agentsSlice, ok := agents.([]interface{})
	if !ok {
		return nil
	}

	result := make([]map[string]interface{}, len(agentsSlice))
	for i, agent := range agentsSlice {
		agentMap, ok := agent.(map[string]interface{})
		if !ok {
			// Handle the error, maybe log it
			continue
		}
		result[i] = agentMap
	}

	return result
}

/* Prompt returns the prompt from the config. */
func Prompt(key string) string { return cfg.v.GetString("ai.setups." + Setup() + ".prompt." + key) }

/* Script returns the script from the config. */
func Script(key string) []string {
	return cfg.v.GetStringSlice("ai.setups." + Setup() + ".script." + key)
}

/* Action returns the action from the config. */
func Action(key string) string {
	return cfg.v.GetString("ai.setups." + Setup() + ".script.actions." + key)
}

/* Flow returns the flow from the config. */
func Flow() []map[string]interface{} {
	flow := cfg.v.Get("ai.setups." + Setup() + ".flow")
	if flow == nil {
		return nil
	}

	// Flow is expected to be a slice of maps (each representing an agent type with instructions).
	flowSlice, ok := flow.([]interface{})
	if !ok {
		return nil
	}

	result := make([]map[string]interface{}, len(flowSlice))
	for i, flowItem := range flowSlice {
		flowMap, ok := flowItem.(map[string]interface{})
		if !ok {
			// Handle the error, maybe log it
			continue
		}
		result[i] = flowMap
	}

	return result
}
