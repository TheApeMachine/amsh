package process

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/berrt"
	"github.com/theapemachine/amsh/errnie"
)

/*
Thinking is a process that allows the system to think about a given topic.
It now includes a detailed reasoning graph to capture multi-level and interconnected reasoning steps.
*/
type Thinking struct {
	HypergraphLayer     HypergraphLayer     `json:"hypergraph_layer" jsonschema:"description:Represents many-to-many relationships and group dynamics,required"`
	TensorNetwork       TensorNetwork       `json:"tensor_network" jsonschema:"description:Multi-dimensional relationship patterns,required"`
	FractalStructure    FractalStructure    `json:"fractal_structure" jsonschema:"description:Self-similar patterns at different scales,required"`
	QuantumLayer        QuantumLayer        `json:"quantum_layer" jsonschema:"description:Probabilistic and superposition states,required"`
	HolographicMemory   HolographicMemory   `json:"holographic_memory" jsonschema:"description:Distributed information storage,required"`
	TemporalDynamics    TemporalDynamics    `json:"temporal_dynamics" jsonschema:"description:Time-based evolution of thoughts,required"`
	EmergentPatterns    EmergentPatterns    `json:"emergent_patterns" jsonschema:"description:Higher-order patterns that emerge from interactions,required"`
	CrossLayerSynthesis CrossLayerSynthesis `json:"cross_layer_synthesis" jsonschema:"description:Integration across different representation layers,required"`
	UnifiedPerspective  UnifiedPerspective  `json:"unified_perspective" jsonschema:"description:Coherent view across all structures,required"`
}

type Conflict struct {
	ID         string   `json:"id" jsonschema:"required,description:Unique identifier for the conflict"`
	Elements   []string `json:"elements" jsonschema:"required,description:Elements in conflict"`
	Type       string   `json:"type" jsonschema:"required,description:Type of conflict"`
	Severity   float64  `json:"severity" jsonschema:"required,description:Severity of the conflict"`
	Resolution string   `json:"resolution" jsonschema:"required,description:Resolution of the conflict"`
}

type UnifiedInsight struct {
	ID           string   `json:"id" jsonschema:"required,description:Unique identifier for the insight"`
	Description  string   `json:"description" jsonschema:"required,description:Description of the insight"`
	Sources      []string `json:"sources" jsonschema:"required,description:Sources of the insight"`
	Confidence   float64  `json:"confidence" jsonschema:"required,description:Confidence in the insight"`
	Impact       float64  `json:"impact" jsonschema:"required,description:Impact of the insight"`
	Applications []string `json:"applications" jsonschema:"required,description:Applications of the insight"`
}

type Event struct {
	ID        string                 `json:"id" jsonschema:"required,description:Unique identifier for event"`
	Type      string                 `json:"type" jsonschema:"required,description:Type of event"`
	Data      map[string]interface{} `json:"data" jsonschema:"description:Event data"`
	Timestamp time.Time              `json:"timestamp" jsonschema:"description:Event time"`
}

// Helper types
type Properties map[string]interface{}

type Metrics struct {
	Coherence  float64 `json:"coherence" jsonschema:"required,description:Coherence metric"`
	Complexity float64 `json:"complexity" jsonschema:"required,description:Complexity metric"`
	Stability  float64 `json:"stability" jsonschema:"required,description:Stability metric"`
	Novelty    float64 `json:"novelty" jsonschema:"required,description:Novelty metric"`
}

type ProcessResult struct {
	CoreID string          `json:"core_id" jsonschema:"required,description:Core ID,"`
	Data   json.RawMessage `json:"data" jsonschema:"description:Data from the core"`
	Error  error           `json:"error" jsonschema:"description:Error from the core"`
}

// Integration type for final results
type ThinkingResult struct {
	Surface SurfaceAnalysis `json:"surface" jsonschema:"required,description:Surface analysis"`
	Pattern PatternAnalysis `json:"pattern" jsonschema:"required,description:Pattern analysis"`
	Quantum QuantumAnalysis `json:"quantum" jsonschema:"required,description:Quantum analysis"`
	Time    TimeAnalysis    `json:"time" jsonschema:"required,description:Time analysis"`
}

func (tr *ThinkingResult) Integrate(result ProcessResult) error {
	var err error
	switch result.CoreID {
	case "surface":
		err = json.Unmarshal(result.Data, &tr.Surface)
	case "pattern":
		err = json.Unmarshal(result.Data, &tr.Pattern)
	case "quantum":
		err = json.Unmarshal(result.Data, &tr.Quantum)
	case "time":
		err = json.Unmarshal(result.Data, &tr.Time)
	default:
		return fmt.Errorf("unknown core ID: %s", result.CoreID)
	}
	return err
}

/*
NewThinking creates a new instance of the Thinking process.
*/
func NewThinking() *Thinking {
	return &Thinking{}
}

/*
Marshal the process into JSON.
*/
func (thinking *Thinking) Marshal() ([]byte, error) {
	return json.Marshal(thinking)
}

/*
Unmarshal the process from JSON.
*/
func (thinking *Thinking) Unmarshal(data []byte) error {
	return json.Unmarshal(data, thinking)
}

func (thinking *Thinking) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.default.prompt", key))
	prompt = strings.ReplaceAll(prompt, "{{schemas}}", thinking.GenerateSchema())

	return prompt
}

func (thinking *Thinking) GenerateSchema() string {
	schema := jsonschema.Reflect(&Thinking{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}

	return string(out)
}

/*
Format the process as a pretty-printed JSON string.
*/
func (thinking *Thinking) Format() string {
	pretty, _ := json.MarshalIndent(thinking, "", "  ")
	return string(pretty)
}

/*
String returns a human-readable string representation of the process.
*/
func (thinking *Thinking) String() {
	berrt.Info("Thinking", thinking)
}
