package process

import "github.com/theapemachine/amsh/utils"

// Core analysis types that implement the Process interface
type SurfaceAnalysis struct {
	HypergraphLayer HypergraphLayer `json:"hypergraph_layer" jsonschema:"description:Represents many-to-many relationships and group dynamics,required"`
	TensorNetwork   TensorNetwork   `json:"tensor_network" jsonschema:"description:Multi-dimensional relationship patterns,required"`
}

func (surface *SurfaceAnalysis) SystemPrompt(key string) string {
	return utils.SystemPrompt(key, "surface", utils.GenerateSchema[SurfaceAnalysis]())
}
