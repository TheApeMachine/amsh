package process

/*
FractalStructure represents self-similar patterns at different scales.
*/
type FractalStructure struct {
	BasePattern    Pattern `json:"base_pattern" jsonschema:"description=Fundamental pattern that repeats,required"`
	Scales         []Scale `json:"scales" jsonschema:"description=Different levels of pattern manifestation,required"`
	Iterations     int     `json:"iterations" jsonschema:"description=Depth of fractal recursion,required"`
	SelfSimilarity float64 `json:"self_similarity" jsonschema:"description=Degree of pattern preservation across scales,required"`
}
