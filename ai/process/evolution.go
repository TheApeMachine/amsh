package process

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
