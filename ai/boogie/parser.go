package boogie

import (
	"fmt"

	"github.com/theapemachine/amsh/errnie"
)

type NodeType int

const (
	NODE_PROGRAM NodeType = iota
	NODE_OPERATION
	NODE_BEHAVIOR
	NODE_FLOW
	NODE_MATCH
	NODE_CLOSURE
	NODE_JOIN
)

type Node struct {
	Type     NodeType
	Value    string
	Behavior *Node
	Next     []*Node
	Parent   *Node
}
type Parser struct {
	current  *Node  // Current node we're building
	program  *Node  // Root of our AST
	lastFlow string // Tracks the last flow operator encountered
}

func NewParser() *Parser {
	program := &Node{
		Type: NODE_PROGRAM,
		Next: make([]*Node, 0),
	}

	return &Parser{
		program:  program,
		current:  program,
		lastFlow: "",
	}
}

func (parser *Parser) Generate(tokens chan Lexeme) *Node {
	errnie.Log("parser.Generate()")

	for token := range tokens {
		//fmt.Printf("Token: %s\n", token.Text)
		switch token.ID {
		case DELIMITER:
			parser.handleDelimiter(token)
		case OPERATION:
			parser.handleOperation(token)
		case FLOW:
			// Skip flow tokens as they don't need to create nodes
			continue
		case VALUE:
			// Skip value tokens (in/out) as they don't need to create nodes
			continue
		default:
			fmt.Printf("Unhandled token: %s\n", token.Text)
		}
	}

	return parser.program
}

func (parser *Parser) handleDelimiter(token Lexeme) {
	errnie.Log("parser.handleDelimiter(%v)", token)

	switch token.Text {
	case "(":
		closure := &Node{
			Type:   NODE_CLOSURE,
			Next:   make([]*Node, 0),
			Parent: parser.current,
		}

		// Add closure to parent's Next slice
		parser.current.Next = append(parser.current.Next, closure)
		parser.current = closure
	case ")":
		parser.current = parser.current.Parent
	}
}

func (parser *Parser) handleOperation(token Lexeme) {
	// Handle join operation specially
	if token.Text == "join" {
		joinNode := &Node{
			Type:   NODE_JOIN,
			Value:  token.Text,
			Next:   make([]*Node, 0),
			Parent: parser.current,
		}
		parser.current.Next = append(parser.current.Next, joinNode)
		parser.current = joinNode
		return
	}

	operation := &Node{
		Type:   NODE_OPERATION,
		Value:  token.Text,
		Next:   make([]*Node, 0),
		Parent: parser.current,
	}
	parser.current.Next = append(parser.current.Next, operation)
}

func (parser *Parser) PrintAST(node *Node, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	fmt.Printf("%s- Type: %v, Value: %s\n", indent, node.Type, node.Value)
	for _, child := range node.Next {
		parser.PrintAST(child, depth+1)
	}
}
