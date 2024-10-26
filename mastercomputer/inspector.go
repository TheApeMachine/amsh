package mastercomputer

type Inspector struct {
}

func NewInspector(parameters map[string]any) *Inspector {
	return &Inspector{}
}

func (inspector *Inspector) Start() string {
	return "Inspector started"
}
