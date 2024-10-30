package process

type CausalChain struct {
	ID       string     `json:"id" jsonschema:"required,description:Unique identifier for causal chain"`
	EventIDs []string   `json:"event_ids" jsonschema:"required,description:IDs of events in chain"`
	Strength float64    `json:"strength" jsonschema:"required,description:Causal relationship strength"`
	Evidence []Evidence `json:"evidence" jsonschema:"required,description:Supporting evidence"`
}

type Evidence struct {
	Type        string  `json:"type" jsonschema:"required,description:Type of evidence"`
	Description string  `json:"description" jsonschema:"required,description:Evidence description"`
	Confidence  float64 `json:"confidence" jsonschema:"required,description:Confidence level"`
	Source      string  `json:"source" jsonschema:"required,description:Evidence source"`
}
