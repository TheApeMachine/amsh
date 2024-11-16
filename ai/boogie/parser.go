package boogie

// AST structures

type Program struct {
	Input  string
	Output string
	Root   *Operation
}

type Operation struct {
	Type       string
	Parameters []string
	Behavior   string
	Label      string // For labels like [flow]
	Outcomes   []string
	Children   []*Operation
}

// Parser struct

type Parser struct {
	tokens  []Lexeme
	current int
}

func NewParser() *Parser {
	return &Parser{}
}

// NewOperation creates a new operation node
func NewOperation(opType string) *Operation {
	return &Operation{
		Type:       opType,
		Parameters: make([]string, 0),
		Behavior:   "",
		Label:      "",
		Outcomes:   make([]string, 0),
		Children:   make([]*Operation, 0),
	}
}
