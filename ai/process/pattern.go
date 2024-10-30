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

type GlobalPattern struct {
	ID           string     `json:"id" jsonschema:"required,description:Unique identifier for the global pattern"`
	Layers       []string   `json:"layers" jsonschema:"required,description:Layers containing the global pattern"`
	Pattern      Pattern    `json:"pattern" jsonschema:"required,description:Pattern of the global pattern"`
	Significance float64    `json:"significance" jsonschema:"required,description:Significance of the global pattern"`
	Support      []Evidence `json:"support" jsonschema:"required,description:Support for the global pattern"`
}

type Scale struct {
	Level      int       `json:"level" jsonschema:"required,description:Level of the scale"`
	Resolution float64   `json:"resolution" jsonschema:"required,description:Resolution of the scale"`
	Patterns   []Pattern `json:"patterns" jsonschema:"required,description:Patterns at this scale"`
	Metrics    Metrics   `json:"metrics" jsonschema:"required,description:Metrics for the scale"`
}

/*
EmergentPatterns represents higher-order patterns that emerge from interactions.
*/
type EmergentPatterns struct {
	Patterns         []Pattern         `json:"patterns" jsonschema:"description:Discovered higher-order patterns; required"`
	EmergenceRules   []EmergenceRule   `json:"emergence_rules" jsonschema:"description:Rules governing pattern formation; required"`
	StabilityMetrics []StabilityMetric `json:"stability_metrics" jsonschema:"description:Measures of pattern stability; required"`
}

type EmergenceRule struct {
	ID           string      `json:"id" jsonschema:"required,description:Unique identifier for the emergence rule"`
	Components   []Pattern   `json:"components" jsonschema:"required,description:Components of the emergence rule"`
	Interactions []Relation  `json:"interactions" jsonschema:"required,description:Interactions between components"`
	Outcome      Pattern     `json:"outcome" jsonschema:"required,description:Outcome of the emergence rule"`
	Conditions   []Predicate `json:"conditions" jsonschema:"required,description:Conditions for the emergence rule"`
}

type StabilityMetric struct {
	Type      string        `json:"type" jsonschema:"required,description:Type of the stability metric"`
	Value     float64       `json:"value" jsonschema:"required,description:Value of the stability metric"`
	Threshold float64       `json:"threshold" jsonschema:"required,description:Threshold for the stability metric"`
	Window    time.Duration `json:"window" jsonschema:"required,description:Window for the stability metric"`
}

/*
UnifiedPerspective represents a coherent view across all structures.
*/
type UnifiedPerspective struct {
	GlobalPatterns []GlobalPattern  `json:"global_patterns" jsonschema:"required,description:Patterns visible across all layers"`
	Coherence      float64          `json:"coherence" jsonschema:"required,description:Measure of overall integration"`
	Insights       []UnifiedInsight `json:"insights" jsonschema:"required,description:Understanding derived from the whole"`
}

type PatternAnalysis struct {
	FractalStructure FractalStructure `json:"fractal_structure" jsonschema:"description:Self-similar patterns at different scales,required"`
	EmergentPatterns EmergentPatterns `json:"emergent_patterns" jsonschema:"description:Higher-order patterns that emerge from interactions,required"`
}

func NewPatternAnalysis() Process {
	return &PatternAnalysis{}
}

// Similar implementations for PatternAnalysis
func (pa *PatternAnalysis) GenerateSchema() string {
	schema := jsonschema.Reflect(&PatternAnalysis{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (pa *PatternAnalysis) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.pattern.prompt", key))
	return strings.ReplaceAll(prompt, "{{schemas}}", pa.GenerateSchema())
}
