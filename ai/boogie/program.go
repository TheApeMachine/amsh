package boogie

// Program represents a pipeline program
type Program struct {
	Type       string // e.g., "switch", "select", "join"
	Operations []Operation
	Input      interface{}
	Output     chan State
}


func NewProgram() *Program {
	return &Program{
		Output: make(chan State),
	}
}
