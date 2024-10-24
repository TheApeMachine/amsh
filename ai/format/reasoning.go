package format

type Reasoning struct {
	Thoughts []Thought `json:"thoughts" jsonschema:"description=Your current thoughts"`
	Done     bool      `json:"done" jsonschema:"description=Set to true when you believe you are done with your current assignment"`
}

func NewReasoning() *Reasoning {
	return &Reasoning{}
}

type Thought struct {
	Chain           []Thought `json:"chain,omitempty" jsonschema:"description=A chain of thoughts"`
	Tree            []Thought `json:"tree,omitempty" jsonschema:"description=A tree of thoughts"`
	Ideas           []Thought `json:"ideas,omitempty" jsonschema:"description=A list of ideas"`
	Realizations    []Thought `json:"realizations,omitempty" jsonschema:"description=Realizations you have about your current situation"`
	SelfReflections []Thought `json:"self_reflections,omitempty" jsonschema:"description=The reflections of the thought"`
}
