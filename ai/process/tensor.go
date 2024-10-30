package process

/*
TensorNetwork represents multi-dimensional relationships and patterns.
*/
type TensorNetwork struct {
	Dimensions   []Dimension   `json:"dimensions" jsonschema:"description:Different aspects of relationship space,required"`
	TensorFields []TensorField `json:"tensor_fields" jsonschema:"description:Multi-dimensional relationship patterns,required"`
	Projections  []Projection  `json:"projections" jsonschema:"description:Lower-dimensional views of the tensor space,required"`
}

type Dimension struct {
	ID         string    `json:"id" jsonschema:"required,description:Unique identifier for the dimension"`
	Name       string    `json:"name" jsonschema:"required,description:Name of the dimension"`
	Scale      []float64 `json:"scale" jsonschema:"required,description:Scale values for the dimension"`
	Resolution float64   `json:"resolution" jsonschema:"required,description:Granularity of the dimension"`
	Boundaries []float64 `json:"boundaries" jsonschema:"required,description:Min and max values"`
}

type TensorField struct {
	ID           string                 `json:"id" jsonschema:"required,description:Unique identifier for the tensor field"`
	DimensionIDs []string               `json:"dimension_ids" jsonschema:"required,description:IDs of dimensions"`
	Values       []float64              `json:"values" jsonschema:"required,description:Flattened tensor values"`
	Shape        []int                  `json:"shape" jsonschema:"required,description:Shape of the tensor"`
	Metadata     map[string]interface{} `json:"metadata" jsonschema:"description:Additional metadata"`
}

type Projection struct {
	ID             string    `json:"id" jsonschema:"required,description:Unique identifier for the projection"`
	SourceDimIDs   []string  `json:"source_dimension_ids" jsonschema:"required,description:IDs of source dimensions"`
	TargetDimIDs   []string  `json:"target_dimension_ids" jsonschema:"required,description:IDs of target dimensions"`
	ProjectionType string    `json:"projection_type" jsonschema:"required,description:Type of projection"`
	Matrix         []float64 `json:"matrix" jsonschema:"required,description:Projection matrix"`
}
