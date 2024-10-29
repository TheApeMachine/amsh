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
	HypergraphLayer   HypergraphLayer   `json:"hypergraph_layer" jsonschema:"description:Represents many-to-many relationships and group dynamics; required"`
	TensorNetwork     TensorNetwork     `json:"tensor_network" jsonschema:"description:Multi-dimensional relationship patterns; required"`
	FractalStructure  FractalStructure  `json:"fractal_structure" jsonschema:"description:Self-similar patterns at different scales; required"`
	QuantumLayer      QuantumLayer      `json:"quantum_layer" jsonschema:"description:Probabilistic and superposition states; required"`
	HolographicMemory HolographicMemory `json:"holographic_memory" jsonschema:"description:Distributed information storage; required"`
	TemporalDynamics  TemporalDynamics  `json:"temporal_dynamics" jsonschema:"description:Time-based evolution of thoughts; required"`
	EmergentPatterns  EmergentPatterns  `json:"emergent_patterns" jsonschema:"description:Higher-order patterns that emerge from interactions; required"`

	// Integration and synthesis
	CrossLayerSynthesis CrossLayerSynthesis `json:"cross_layer_synthesis" jsonschema:"description:Integration across different representation layers; required"`
	UnifiedPerspective  UnifiedPerspective  `json:"unified_perspective" jsonschema:"description:Coherent view across all structures; required"`
}

/*
HypergraphLayer represents relationships between multiple nodes simultaneously.
*/
type HypergraphLayer struct {
	Nodes      []HyperNode   `json:"nodes" jsonschema:"description:Nodes in the hypergraph; required"`
	HyperEdges []HyperEdge   `json:"hyper_edges" jsonschema:"description:Edges connecting multiple nodes; required"`
	Clusters   []NodeCluster `json:"clusters" jsonschema:"description:Emergent groupings of nodes; required"`
}

/*
TensorNetwork represents multi-dimensional relationships and patterns.
*/
type TensorNetwork struct {
	Dimensions   []Dimension   `json:"dimensions" jsonschema:"description:Different aspects of relationship space; required"`
	TensorFields []TensorField `json:"tensor_fields" jsonschema:"description:Multi-dimensional relationship patterns; required"`
	Projections  []Projection  `json:"projections" jsonschema:"description:Lower-dimensional views of the tensor space; required"`
}

/*
FractalStructure represents self-similar patterns at different scales.
*/
type FractalStructure struct {
	BasePattern    Pattern `json:"base_pattern" jsonschema:"description:Fundamental pattern that repeats; required"`
	Scales         []Scale `json:"scales" jsonschema:"description:Different levels of pattern manifestation; required"`
	Iterations     int     `json:"iterations" jsonschema:"description:Depth of fractal recursion; required"`
	SelfSimilarity float64 `json:"self_similarity" jsonschema:"description:Degree of pattern preservation across scales; required"`
}

/*
QuantumLayer represents probabilistic and superposition states of thoughts.
*/
type QuantumLayer struct {
	SuperpositionStates []SuperpositionState `json:"superposition_states" jsonschema:"description:Multiple simultaneous possibilities; required"`
	Entanglements       []Entanglement       `json:"entanglements" jsonschema:"description:Correlated state relationships; required"`
	WaveFunction        WaveFunction         `json:"wave_function" jsonschema:"description:Probability distribution of states; required"`
}

/*
HolographicMemory represents distributed information storage.
*/
type HolographicMemory struct {
	Encodings          []Encoding        `json:"encodings" jsonschema:"description:Distributed information patterns; required"`
	InterferenceSpace  InterferenceSpace `json:"interference_space" jsonschema:"description:Interaction between encodings; required"`
	ReconstructionKeys []string          `json:"reconstruction_keys" jsonschema:"description:Access patterns for information retrieval; required"`
}

/*
TemporalDynamics represents the evolution of thoughts over time.
*/
type TemporalDynamics struct {
	Timeline       []TimePoint     `json:"timeline" jsonschema:"description:Sequence of thought states; required"`
	CausalChains   []CausalChain   `json:"causal_chains" jsonschema:"description:Cause-effect relationships over time; required"`
	EvolutionRules []EvolutionRule `json:"evolution_rules" jsonschema:"description:Patterns of state change; required"`
}

/*
EmergentPatterns represents higher-order patterns that emerge from interactions.
*/
type EmergentPatterns struct {
	Patterns         []Pattern         `json:"patterns" jsonschema:"description:Discovered higher-order patterns; required"`
	EmergenceRules   []EmergenceRule   `json:"emergence_rules" jsonschema:"description:Rules governing pattern formation; required"`
	StabilityMetrics []StabilityMetric `json:"stability_metrics" jsonschema:"description:Measures of pattern stability; required"`
}

/*
CrossLayerSynthesis represents integration across different representation layers.
*/
type CrossLayerSynthesis struct {
	Mappings     []LayerMapping `json:"mappings" jsonschema:"description:Correspondences between layers; required"`
	Integrations []Integration  `json:"integrations" jsonschema:"description:Unified patterns across layers; required"`
	Conflicts    []Conflict     `json:"conflicts" jsonschema:"description:Contradictions between layers; required"`
}

/*
UnifiedPerspective represents a coherent view across all structures.
*/
type UnifiedPerspective struct {
	GlobalPatterns []GlobalPattern  `json:"global_patterns" jsonschema:"description:Patterns visible across all layers; required"`
	Coherence      float64          `json:"coherence" jsonschema:"description:Measure of overall integration; required"`
	Insights       []UnifiedInsight `json:"insights" jsonschema:"description:Understanding derived from the whole; required"`
}

type Projection struct {
	ID             string    `json:"id" jsonschema:"required,description:Unique identifier for the projection"`
	SourceDimIDs   []string  `json:"source_dimension_ids" jsonschema:"required,description:IDs of source dimensions"`
	TargetDimIDs   []string  `json:"target_dimension_ids" jsonschema:"required,description:IDs of target dimensions"`
	ProjectionType string    `json:"projection_type" jsonschema:"required,description:Type of projection"`
	Matrix         []float64 `json:"matrix" jsonschema:"required,description:Projection matrix"`
}

type Scale struct {
	Level      int       `json:"level" jsonschema:"required,description:Level of the scale"`
	Resolution float64   `json:"resolution" jsonschema:"required,description:Resolution of the scale"`
	Patterns   []Pattern `json:"patterns" jsonschema:"required,description:Patterns at this scale"`
	Metrics    Metrics   `json:"metrics" jsonschema:"required,description:Metrics for the scale"`
}

type Encoding struct {
	ID       string    `json:"id" jsonschema:"required,description:Unique identifier for the encoding"`
	Pattern  []float64 `json:"pattern" jsonschema:"required,description:Pattern of the encoding"`
	Phase    float64   `json:"phase" jsonschema:"required,description:Phase of the encoding"`
	Position []int     `json:"position" jsonschema:"required,description:Position of the encoding"`
	Strength float64   `json:"strength" jsonschema:"required,description:Strength of the encoding"`
}

type InterferenceSpace struct {
	Dimensions []int       `json:"dimensions" jsonschema:"required,description:Dimensions of the interference space"`
	Field      []float64   `json:"field" jsonschema:"required,description:Field of the interference space"`
	Resonances []Resonance `json:"resonances" jsonschema:"required,description:Resonances in the interference space"`
	Energy     float64     `json:"energy" jsonschema:"required,description:Energy of the interference space"`
}

type Resonance struct {
	Position  []int   `json:"position" jsonschema:"required,description:Position of the resonance"`
	Strength  float64 `json:"strength" jsonschema:"required,description:Strength of the resonance"`
	Phase     float64 `json:"phase" jsonschema:"required,description:Phase of the resonance"`
	Stability float64 `json:"stability" jsonschema:"required,description:Stability of the resonance"`
}

type EvolutionRule struct {
	ID          string    `json:"id" jsonschema:"required,description:Unique identifier for the evolution rule"`
	Condition   Predicate `json:"condition" jsonschema:"required,description:Condition for the evolution rule"`
	Action      Transform `json:"action" jsonschema:"required,description:Action to be taken"`
	Priority    int       `json:"priority" jsonschema:"required,description:Priority of the evolution rule"`
	Reliability float64   `json:"reliability" jsonschema:"required,description:Reliability of the evolution rule"`
}

type Predicate struct {
	Type      string                 `json:"type" jsonschema:"required,description:Type of the predicate"`
	Params    map[string]interface{} `json:"params" jsonschema:"required,description:Parameters for the predicate"`
	Threshold float64                `json:"threshold" jsonschema:"required,description:Threshold for the predicate"`
}

type Transform struct {
	Type      string                 `json:"type" jsonschema:"required,description:Type of the transformation"`
	Params    map[string]interface{} `json:"params" jsonschema:"required,description:Parameters for the transformation"`
	Magnitude float64                `json:"magnitude" jsonschema:"required,description:Magnitude of the transformation"`
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

type LayerMapping struct {
	FromLayer  string    `json:"from_layer" jsonschema:"required,description:Source layer"`
	ToLayer    string    `json:"to_layer" jsonschema:"required,description:Target layer"`
	Mappings   []Mapping `json:"mappings" jsonschema:"required,description:Mappings between layers"`
	Confidence float64   `json:"confidence" jsonschema:"required,description:Confidence in the layer mapping"`
}

type Mapping struct {
	FromID string                 `json:"from_id" jsonschema:"required,description:Source ID"`
	ToID   string                 `json:"to_id" jsonschema:"required,description:Target ID"`
	Type   string                 `json:"type" jsonschema:"required,description:Type of mapping"`
	Params map[string]interface{} `json:"params" jsonschema:"required,description:Parameters for the mapping"`
}

type Integration struct {
	ID        string    `json:"id" jsonschema:"required,description:Unique identifier for the integration"`
	Patterns  []Pattern `json:"patterns" jsonschema:"required,description:Patterns integrated"`
	Mappings  []Mapping `json:"mappings" jsonschema:"required,description:Mappings between patterns"`
	Coherence float64   `json:"coherence" jsonschema:"required,description:Coherence of the integration"`
	Stability float64   `json:"stability" jsonschema:"required,description:Stability of the integration"`
}

type Conflict struct {
	ID         string   `json:"id" jsonschema:"required,description:Unique identifier for the conflict"`
	Elements   []string `json:"elements" jsonschema:"required,description:Elements in conflict"`
	Type       string   `json:"type" jsonschema:"required,description:Type of conflict"`
	Severity   float64  `json:"severity" jsonschema:"required,description:Severity of the conflict"`
	Resolution string   `json:"resolution" jsonschema:"required,description:Resolution of the conflict"`
}

type GlobalPattern struct {
	ID           string     `json:"id" jsonschema:"required,description:Unique identifier for the global pattern"`
	Layers       []string   `json:"layers" jsonschema:"required,description:Layers containing the global pattern"`
	Pattern      Pattern    `json:"pattern" jsonschema:"required,description:Pattern of the global pattern"`
	Significance float64    `json:"significance" jsonschema:"required,description:Significance of the global pattern"`
	Support      []Evidence `json:"support" jsonschema:"required,description:Support for the global pattern"`
}

type UnifiedInsight struct {
	ID           string   `json:"id" jsonschema:"required,description:Unique identifier for the insight"`
	Description  string   `json:"description" jsonschema:"required,description:Description of the insight"`
	Sources      []string `json:"sources" jsonschema:"required,description:Sources of the insight"`
	Confidence   float64  `json:"confidence" jsonschema:"required,description:Confidence in the insight"`
	Impact       float64  `json:"impact" jsonschema:"required,description:Impact of the insight"`
	Applications []string `json:"applications" jsonschema:"required,description:Applications of the insight"`
}

// ComplexNumber represents a complex number in JSON-serializable format
type ComplexNumber struct {
	Real      float64 `json:"real" jsonschema:"required,description:Real part of the complex number"`
	Imaginary float64 `json:"imaginary" jsonschema:"required,description:Imaginary part of the complex number"`
}

// WaveFunction now uses JSON-serializable complex numbers
type WaveFunction struct {
	StateSpaceDim int             `json:"state_space_dim" jsonschema:"description:Dimension of the state space"`
	Amplitudes    []ComplexNumber `json:"amplitudes" jsonschema:"description:Quantum state amplitudes"`
	Basis         []string        `json:"basis" jsonschema:"description:Names of basis states"`
	Time          time.Time       `json:"time" jsonschema:"description:Time of the wave function"`
}

type HyperNode struct {
	ID         string                 `json:"id" jsonschema:"required,description:Unique identifier for the node"`
	Content    interface{}            `json:"content" jsonschema:"description:The content or value of the node"`
	Properties map[string]interface{} `json:"properties" jsonschema:"description:Additional properties of the node"`
	Dimension  int                    `json:"dimension" jsonschema:"description:Dimensionality of the node"`
	Weight     float64                `json:"weight" jsonschema:"description:Importance or strength of the node"`
	Activation float64                `json:"activation" jsonschema:"description:Current activation level"`
}

type HyperEdge struct {
	ID       string   `json:"id" jsonschema:"required,description:Unique identifier for the edge"`
	NodeIDs  []string `json:"node_ids" jsonschema:"required,description:IDs of connected nodes"`
	Type     string   `json:"type" jsonschema:"description:Type of relationship"`
	Strength float64  `json:"strength" jsonschema:"description:Strength of the connection"`
	Context  string   `json:"context" jsonschema:"description:Context of the relationship"`
}

type NodeCluster struct {
	ID        string    `json:"id" jsonschema:"required,description:Unique identifier for the cluster"`
	NodeIDs   []string  `json:"node_ids" jsonschema:"required,description:IDs of nodes in the cluster"`
	Centroid  []float64 `json:"centroid" jsonschema:"description:Center point of the cluster"`
	Coherence float64   `json:"coherence" jsonschema:"description:Measure of cluster coherence"`
	Label     string    `json:"label" jsonschema:"description:Descriptive label for the cluster"`
}

type Dimension struct {
	ID         string    `json:"id" jsonschema:"required,description:Unique identifier for the dimension"`
	Name       string    `json:"name" jsonschema:"required,description:Name of the dimension"`
	Scale      []float64 `json:"scale" jsonschema:"description:Scale values for the dimension"`
	Resolution float64   `json:"resolution" jsonschema:"description:Granularity of the dimension"`
	Boundaries []float64 `json:"boundaries" jsonschema:"description:Min and max values"`
}

type TensorField struct {
	ID           string                 `json:"id" jsonschema:"required,description:Unique identifier for the tensor field"`
	DimensionIDs []string               `json:"dimension_ids" jsonschema:"required,description:IDs of dimensions"`
	Values       []float64              `json:"values" jsonschema:"description:Flattened tensor values"`
	Shape        []int                  `json:"shape" jsonschema:"description:Shape of the tensor"`
	Metadata     map[string]interface{} `json:"metadata" jsonschema:"description:Additional metadata"`
}

type Pattern struct {
	ID         string     `json:"id" jsonschema:"required,description:Unique identifier for the pattern"`
	Elements   []Element  `json:"elements" jsonschema:"description:Component elements"`
	Relations  []Relation `json:"relations" jsonschema:"description:Relationships between elements"`
	Frequency  float64    `json:"frequency" jsonschema:"description:Occurrence frequency"`
	Confidence float64    `json:"confidence" jsonschema:"description:Confidence in the pattern"`
}

type Element struct {
	ID       string                 `json:"id" jsonschema:"required,description:Unique identifier for the element"`
	Type     string                 `json:"type" jsonschema:"required,description:Type of element"`
	Value    interface{}            `json:"value" jsonschema:"description:Element value"`
	Features map[string]interface{} `json:"features" jsonschema:"description:Element features"`
}

type Relation struct {
	FromID   string                 `json:"from_id" jsonschema:"required,description:Source element ID"`
	ToID     string                 `json:"to_id" jsonschema:"required,description:Target element ID"`
	Type     string                 `json:"type" jsonschema:"required,description:Type of relationship"`
	Weight   float64                `json:"weight" jsonschema:"description:Relationship strength"`
	Metadata map[string]interface{} `json:"metadata" jsonschema:"description:Additional metadata"`
}

type SuperpositionState struct {
	ID            string             `json:"id" jsonschema:"required,description:Unique identifier for the state"`
	Possibilities map[string]float64 `json:"possibilities" jsonschema:"required,description:Possible states and probabilities"`
	Phase         float64            `json:"phase" jsonschema:"description:Quantum phase"`
	Coherence     float64            `json:"coherence" jsonschema:"description:State coherence"`
	Lifetime      time.Duration      `json:"lifetime" jsonschema:"description:Expected lifetime"`
}

type Entanglement struct {
	ID       string        `json:"id" jsonschema:"required,description:Unique identifier for entanglement"`
	StateIDs []string      `json:"state_ids" jsonschema:"required,description:IDs of entangled states"`
	Strength float64       `json:"strength" jsonschema:"description:Entanglement strength"`
	Type     string        `json:"type" jsonschema:"description:Type of entanglement"`
	Duration time.Duration `json:"duration" jsonschema:"description:Expected duration"`
}

type TimePoint struct {
	Time   time.Time              `json:"time" jsonschema:"required,description:Point in time"`
	State  map[string]interface{} `json:"state" jsonschema:"description:System state"`
	Delta  map[string]float64     `json:"delta" jsonschema:"description:State changes"`
	Events []Event                `json:"events" jsonschema:"description:Events at this time"`
}

type Event struct {
	ID        string                 `json:"id" jsonschema:"required,description:Unique identifier for event"`
	Type      string                 `json:"type" jsonschema:"required,description:Type of event"`
	Data      map[string]interface{} `json:"data" jsonschema:"description:Event data"`
	Timestamp time.Time              `json:"timestamp" jsonschema:"description:Event time"`
}

type CausalChain struct {
	ID       string     `json:"id" jsonschema:"required,description:Unique identifier for causal chain"`
	EventIDs []string   `json:"event_ids" jsonschema:"required,description:IDs of events in chain"`
	Strength float64    `json:"strength" jsonschema:"description:Causal relationship strength"`
	Evidence []Evidence `json:"evidence" jsonschema:"description:Supporting evidence"`
}

type Evidence struct {
	Type        string  `json:"type" jsonschema:"required,description:Type of evidence"`
	Description string  `json:"description" jsonschema:"required,description:Evidence description"`
	Confidence  float64 `json:"confidence" jsonschema:"description:Confidence level"`
	Source      string  `json:"source" jsonschema:"description:Evidence source"`
}

// Helper types
type Properties map[string]interface{}

type Metrics struct {
	Coherence  float64 `json:"coherence" jsonschema:"required,description:Coherence metric"`
	Complexity float64 `json:"complexity" jsonschema:"required,description:Complexity metric"`
	Stability  float64 `json:"stability" jsonschema:"required,description:Stability metric"`
	Novelty    float64 `json:"novelty" jsonschema:"required,description:Novelty metric"`
}

// Core analysis types that implement the Process interface
type SurfaceAnalysis struct {
	HypergraphLayer HypergraphLayer `json:"hypergraph_layer" jsonschema:"description:Represents many-to-many relationships and group dynamics; required"`
	TensorNetwork   TensorNetwork   `json:"tensor_network" jsonschema:"description:Multi-dimensional relationship patterns; required"`
}

type PatternAnalysis struct {
	FractalStructure FractalStructure `json:"fractal_structure" jsonschema:"description:Self-similar patterns at different scales; required"`
	EmergentPatterns EmergentPatterns `json:"emergent_patterns" jsonschema:"description:Higher-order patterns that emerge from interactions; required"`
}

type QuantumAnalysis struct {
	QuantumLayer      QuantumLayer      `json:"quantum_layer" jsonschema:"description:Probabilistic and superposition states; required"`
	HolographicMemory HolographicMemory `json:"holographic_memory" jsonschema:"description:Distributed information storage; required"`
}

type TimeAnalysis struct {
	TemporalDynamics    TemporalDynamics    `json:"temporal_dynamics" jsonschema:"description:Time-based evolution of thoughts; required"`
	CrossLayerSynthesis CrossLayerSynthesis `json:"cross_layer_synthesis" jsonschema:"description:Integration across different representation layers; required"`
}

// Constructor functions
func NewSurfaceAnalysis() Process {
	return &SurfaceAnalysis{}
}

func NewPatternAnalysis() Process {
	return &PatternAnalysis{}
}

func NewQuantumAnalysis() Process {
	return &QuantumAnalysis{}
}

func NewTimeAnalysis() Process {
	return &TimeAnalysis{}
}

type ProcessResult struct {
	CoreID string
	Data   json.RawMessage
	Error  error
}

// Process interface implementations for SurfaceAnalysis
func (sa *SurfaceAnalysis) GenerateSchema() string {
	schema := jsonschema.Reflect(&SurfaceAnalysis{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (sa *SurfaceAnalysis) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.surface.prompt", key))
	return strings.ReplaceAll(prompt, "{{schemas}}", sa.GenerateSchema())
}

func (sa *SurfaceAnalysis) Marshal() ([]byte, error) {
	return json.Marshal(sa)
}

func (sa *SurfaceAnalysis) Unmarshal(data []byte) error {
	return json.Unmarshal(data, sa)
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

// Similar implementations for QuantumAnalysis
func (qa *QuantumAnalysis) GenerateSchema() string {
	schema := jsonschema.Reflect(&QuantumAnalysis{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (qa *QuantumAnalysis) SystemPrompt(key string) string {
	prompt := viper.GetViper().GetString(fmt.Sprintf("ai.setups.%s.processes.quantum.prompt", key))
	return strings.ReplaceAll(prompt, "{{schemas}}", qa.GenerateSchema())
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

// Integration type for final results
type ThinkingResult struct {
	Surface *SurfaceAnalysis `json:"surface,omitempty"`
	Pattern *PatternAnalysis `json:"pattern,omitempty"`
	Quantum *QuantumAnalysis `json:"quantum,omitempty"`
	Time    *TimeAnalysis    `json:"time,omitempty"`
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
