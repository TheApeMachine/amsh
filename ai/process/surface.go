package process

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

// Core analysis types that implement the Process interface
type SurfaceAnalysis struct {
	HypergraphLayer HypergraphLayer `json:"hypergraph_layer" jsonschema:"description:Represents many-to-many relationships and group dynamics,required"`
	TensorNetwork   TensorNetwork   `json:"tensor_network" jsonschema:"description:Multi-dimensional relationship patterns,required"`
}

func NewSurfaceAnalysis() Process {
	return &SurfaceAnalysis{}
}

/*
Analyze implements the Analysis interface for SurfaceAnalysis.
*/
func (analysis *SurfaceAnalysis) Analyze(input string) (interface{}, error) {
	// Implement surface analysis logic
	return nil, nil
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
