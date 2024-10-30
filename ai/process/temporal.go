package process

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
TemporalDynamics represents the evolution of thoughts over time.
*/
type TemporalDynamics struct {
	Timeline       []TimePoint     `json:"timeline" jsonschema:"description:Sequence of thought states; required"`
	CausalChains   []CausalChain   `json:"causal_chains" jsonschema:"description:Cause-effect relationships over time; required"`
	EvolutionRules []EvolutionRule `json:"evolution_rules" jsonschema:"description:Patterns of state change; required"`
}

type TimePoint struct {
	Time   time.Time              `json:"time" jsonschema:"required,description:Point in time"`
	State  map[string]interface{} `json:"state" jsonschema:"required,description:System state"`
	Delta  map[string]float64     `json:"delta" jsonschema:"required,description:State changes"`
	Events []Event                `json:"events" jsonschema:"required,description:Events at this time"`
}

type TimeAnalysis struct {
	TemporalDynamics    TemporalDynamics    `json:"temporal_dynamics" jsonschema:"description:Time-based evolution of thoughts,required"`
	CrossLayerSynthesis CrossLayerSynthesis `json:"cross_layer_synthesis" jsonschema:"description:Integration across different representation layers,required"`
}

func NewTimeAnalysis() Process {
	return &TimeAnalysis{}
}

// Similar implementations for TimeAnalysis
func (ta *TimeAnalysis) GenerateSchema() string {
	schema := jsonschema.Reflect(&TimeAnalysis{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (ta *TimeAnalysis) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.time.prompt", key))
	return strings.ReplaceAll(prompt, "{{schemas}}", ta.GenerateSchema())
}
