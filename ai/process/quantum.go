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
QuantumLayer represents probabilistic and superposition states of thoughts.
*/
type QuantumLayer struct {
	SuperpositionStates []SuperpositionState `json:"superposition_states" jsonschema:"required,description:Multiple simultaneous possibilities"`
	Entanglements       []Entanglement       `json:"entanglements" jsonschema:"required,description:Correlated state relationships"`
	WaveFunction        WaveFunction         `json:"wave_function" jsonschema:"required,description:Probability distribution of states"`
}

// ComplexNumber represents a complex number in JSON-serializable format
type ComplexNumber struct {
	Real      float64 `json:"real" jsonschema:"required,description:Real part of the complex number"`
	Imaginary float64 `json:"imaginary" jsonschema:"required,description:Imaginary part of the complex number"`
}

// WaveFunction now uses JSON-serializable complex numbers
type WaveFunction struct {
	StateSpaceDim int             `json:"state_space_dim" jsonschema:"required,description:Dimension of the state space"`
	Amplitudes    []ComplexNumber `json:"amplitudes" jsonschema:"required,description:Quantum state amplitudes"`
	Basis         []string        `json:"basis" jsonschema:"required,description:Names of basis states"`
	Time          time.Time       `json:"time" jsonschema:"required,description:Time of the wave function"`
}

type SuperpositionState struct {
	ID            string             `json:"id" jsonschema:"required,description:Unique identifier for the state"`
	Possibilities map[string]float64 `json:"possibilities" jsonschema:"required,description:Possible states and probabilities"`
	Phase         float64            `json:"phase" jsonschema:"description:Quantum phase"`
	Coherence     float64            `json:"coherence" jsonschema:"required,description:State coherence"`
	Lifetime      time.Duration      `json:"lifetime" jsonschema:"required,description:Expected lifetime"`
}

type Entanglement struct {
	ID       string        `json:"id" jsonschema:"required,description:Unique identifier for entanglement"`
	StateIDs []string      `json:"state_ids" jsonschema:"required,description:IDs of entangled states"`
	Strength float64       `json:"strength" jsonschema:"required,description:Entanglement strength"`
	Type     string        `json:"type" jsonschema:"required,description:Type of entanglement"`
	Duration time.Duration `json:"duration" jsonschema:"required,description:Expected duration"`
}

type QuantumAnalysis struct {
	QuantumLayer      QuantumLayer      `json:"quantum_layer" jsonschema:"description:Probabilistic and superposition states,required"`
	HolographicMemory HolographicMemory `json:"holographic_memory" jsonschema:"description:Distributed information storage,required"`
}

func NewQuantumAnalysis() Process {
	return &QuantumAnalysis{}
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
