package fractal

import (
	"time"

	"github.com/theapemachine/amsh/utils"
)

/*
Pattern encapsulates the fundamental and recurring motifs or structures
that form the basis of the abstract thought or reasoning process.
Each Pattern represents the essential characteristics that persist
across different manifestations, ensuring consistency within the fractal system.
*/
type Pattern struct {
	Identifier  string    `json:"identifier" jsonschema:"description=Unique identifier for the pattern,required"`
	Motif       string    `json:"motif" jsonschema:"description=Motif of the pattern,required"`
	Description string    `json:"description" jsonschema:"description=Detailed description of the pattern,required"`
	CreatedAt   time.Time `json:"created_at" jsonschema:"description=Timestamp of pattern creation,required"`
}

/*
Scale represents the different levels or magnitudes at which the fundamental
patterns are expressed and manifested. Each Scale defines the degree of
abstraction or granularity, allowing patterns to be observed and analyzed
at varying levels of detail.
*/
type Scale struct {
	Level       int     `json:"level" jsonschema:"description=Hierarchical level of the scale,required"`
	Magnitude   float64 `json:"magnitude" jsonschema:"description=Extent of pattern manifestation at this scale,required"`
	Description string  `json:"description" jsonschema:"description=Explanation of the pattern's behavior at this scale,required"`
}

/*
Process represents a fractal thought or reasoning process.
*/
type Process struct {
	BasePattern    Pattern `json:"base_pattern" jsonschema:"description=Fundamental pattern that repeats,required"`
	Scales         []Scale `json:"scales" jsonschema:"description=Different levels of pattern manifestation,required"`
	Iterations     int     `json:"iterations" jsonschema:"description=Depth of fractal recursion,required"`
	SelfSimilarity float64 `json:"self_similarity" jsonschema:"description=Degree of pattern preservation across scales,required"`
}

func (ta *Process) GenerateSchema() string {
	return utils.GenerateSchema[Process]()
}
