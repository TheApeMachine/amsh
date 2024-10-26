package mastercomputer

type Assignment struct {
}

func NewAssignment(parameters map[string]any) *Assignment {
	return &Assignment{}
}

func (assignment *Assignment) Start() string {
	return "Assignment started"
}
