package boogie

import (
	"fmt"
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
	tokens    chan Lexeme
	current   *Node   // Current node we're building
	program   *Node   // Root of our AST
	openNodes []*Node // Stack of open nodes for nested structures
	lastFlow  string  // Tracks the last flow operator encountered
}

func NewParser(tokens chan Lexeme) *Parser {
	program := &Node{
		Type: NODE_PROGRAM,
		Next: make([]*Node, 0),
	}

	return &Parser{
		tokens:    tokens,
		program:   program,
		current:   program,
		openNodes: []*Node{program},
		lastFlow:  "",
	}
}

func (parser *Parser) Generate() *Node {
	for token := range parser.tokens {
		fmt.Printf("Token: %s\n", token.Text)
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
	switch token.Text {
	case "(":
		closure := &Node{
			Type:   NODE_CLOSURE,
			Next:   make([]*Node, 0),
			Parent: parser.current,
		}

		// If parent is a join node, add to its Next slice
		if parser.current.Type == NODE_JOIN {
			parser.current.Next = append(parser.current.Next, closure)
		} else {
			parser.current.Next = append(parser.current.Next, closure)
		}

		parser.openNodes = append(parser.openNodes, closure)
		parser.current = closure
	case ")":
		if len(parser.openNodes) > 0 {
			lastIndex := len(parser.openNodes) - 1
			parser.openNodes = parser.openNodes[:lastIndex] // Pop from stack

			if lastIndex > 0 {
				// If parent is a join node, stay at the join node level
				parent := parser.openNodes[lastIndex-1]
				if parent.Type == NODE_JOIN {
					parser.current = parent
				} else if parent.Parent != nil && parent.Parent.Type == NODE_JOIN {
					parser.current = parent.Parent
				} else {
					parser.current = parent
				}
			} else {
				parser.current = parser.program
			}
		}
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
