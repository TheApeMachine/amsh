package process

import (
	"time"

	"github.com/theapemachine/amsh/utils"
)

/*
TemporalDynamics represents the evolution of thoughts over time.
*/
type TemporalDynamics struct {
	Timeline       []TimePoint     `json:"timeline" jsonschema:"description:Sequence of thought states; required"`
	CausalChains   []CausalChain   `json:"causal_chains" jsonschema:"description:Cause-effect relationships over time; required"`
	EvolutionRules []EvolutionRule `json:"evolution_rules" jsonschema:"description:Patterns of state change; required"`
}

type CausalChain struct {
	ID       string     `json:"id" jsonschema:"required,description:Unique identifier for causal chain"`
	EventIDs []string   `json:"event_ids" jsonschema:"required,description:IDs of events in chain"`
	Strength float64    `json:"strength" jsonschema:"required,description:Causal relationship strength"`
	Evidence []Evidence `json:"evidence" jsonschema:"required,description:Supporting evidence"`
}

type Evidence struct {
	Type        string  `json:"type" jsonschema:"required,description:Type of evidence"`
	Description string  `json:"description" jsonschema:"required,description:Evidence description"`
	Confidence  float64 `json:"confidence" jsonschema:"required,description:Confidence level"`
	Source      string  `json:"source" jsonschema:"required,description:Evidence source"`
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

func (ta *TimeAnalysis) SystemPrompt(key string) string {
	return utils.SystemPrompt(key, "time", utils.GenerateSchema[TimeAnalysis]())
}
