package boogie

import (
	"fmt"
	"strings"
)

// AST structures

type Program struct {
	Input  string
	Output string
	Root   *Operation
}

type Operation struct {
	Type       string
	Behavior   string
	Parameters []string
	Outcomes   []string
	Children   []*Operation
	Label      string
}

// Parser struct

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) Parse() (*Program, error) {
	program := &Program{}

	// Parse 'out <= '
	if err := p.expect(IDENTIFIER, "out"); err != nil {
		return nil, err
	}
	program.Output = "out"

	if err := p.expect(ARROW, "<="); err != nil {
		return nil, err
	}

	// Parse the root operation
	rootOp, err := p.parseOperation()
	if err != nil {
		return nil, err
	}
	program.Root = rootOp

	// Parse ' <= in'
	if err := p.expect(ARROW, "<="); err != nil {
		return nil, err
	}

	if err := p.expect(IDENTIFIER, "in"); err != nil {
		return nil, err
	}
	program.Input = "in"

	return program, nil
}

func (p *Parser) parseOperation() (*Operation, error) {
	// Debugging: Print current token
	fmt.Printf("Parsing operation at position %d: %v\n", p.current, p.currentToken())

	// Handle control flow structures
	if p.check(SWITCH, "") || p.check(SELECT, "") ||
		p.check(MATCH, "") || p.check(JOIN, "") {
		op := &Operation{
			Type:     p.currentToken().Value,
			Children: []*Operation{},
		}
		p.advance()

		// Handle labels for switch and select
		if (op.Type == "switch" || op.Type == "select") && p.check(LABEL, "") {
			op.Label = p.currentToken().Value
			p.advance()
		}

		// Expect arrow
		if err := p.expect(ARROW, "<="); err != nil {
			return nil, err
		}

		// Expect opening delimiter
		if err := p.expect(DELIMITER, "("); err != nil {
			return nil, err
		}

		// Parse children
		for !p.check(DELIMITER, ")") && !p.isAtEnd() {
			childOp, err := p.parseOperation()
			if err != nil {
				return nil, err
			}

			op.Children = append(op.Children, childOp)
		}

		// Expect closing delimiter
		if err := p.expect(DELIMITER, ")"); err != nil {
			return nil, err
		}

		// Validate control structure
		if len(op.Children) == 0 {
			return nil, fmt.Errorf("Expected at least one operation in %s block", op.Type)
		}

		return op, nil
	}

	// Handle regular operations
	op := &Operation{}

	// Check for parameters
	if p.check(PARAMETER, "") {
		op.Parameters = parseParameters(p.currentToken().Value)
		p.advance()
	}

	// Operation type
	if p.check(IDENTIFIER, "") {
		op.Type = p.currentToken().Value
		p.advance()
	} else {
		return nil, fmt.Errorf("Expected operation type at position %d", p.current)
	}

	// Behavior
	if p.check(BEHAVIOR, "") {
		op.Behavior = p.currentToken().Value
		p.advance()
	}

	// Arrow
	if err := p.expect(ARROW, "=>"); err != nil {
		return nil, err
	}

	// Handle nested operations
	if p.check(DELIMITER, "(") {
		p.advance()
		for !p.check(DELIMITER, ")") && !p.isAtEnd() {
			childOp, err := p.parseOperation()
			if err != nil {
				return nil, err
			}
			op.Children = append(op.Children, childOp)
		}
		if err := p.expect(DELIMITER, ")"); err != nil {
			return nil, err
		}
		return op, nil
	}

	// Outcomes
	outcomes, err := p.parseOutcomes()
	if err != nil {
		return nil, err
	}
	op.Outcomes = outcomes

	return op, nil
}

func (p *Parser) parseOutcomes() ([]string, error) {
	outcomes := []string{}

	if p.check(OUTCOME, "") || p.check(IDENTIFIER, "") {
		outcomes = append(outcomes, p.currentToken().Value)
		p.advance()
	} else {
		return nil, fmt.Errorf("Expected outcome at position %d", p.current)
	}

	for p.check(OPERATOR, "|") {
		p.advance()
		if p.check(OUTCOME, "") || p.check(IDENTIFIER, "") {
			outcomes = append(outcomes, p.currentToken().Value)
			p.advance()
		} else {
			return nil, fmt.Errorf("Expected outcome after '|' at position %d", p.current)
		}
	}

	return outcomes, nil
}

func parseParameters(paramStr string) []string {
	raw := paramStr[1 : len(paramStr)-1] // Remove '[' and ']'
	parts := strings.Split(raw, ",")
	params := []string{}
	for _, part := range parts {
		params = append(params, strings.TrimSpace(part))
	}
	return params
}

// Utility methods

func (p *Parser) expect(tokenType TokenType, value string) error {
	if p.check(tokenType, value) {
		p.advance()
		return nil
	}
	return fmt.Errorf("Expected %s '%s' at position %d", tokenType.String(), value, p.current)
}

func (p *Parser) match(tokenType TokenType, value string) bool {
	if p.check(tokenType, value) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) check(tokenType TokenType, value string) bool {
	if p.isAtEnd() {
		return false
	}
	token := p.tokens[p.current]
	if token.Type != tokenType {
		return false
	}
	if value != "" && token.Value != value {
		return false
	}
	return true
}

func (p *Parser) advance() {
	if !p.isAtEnd() {
		p.current++
	}
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) currentToken() Token {
	if p.isAtEnd() {
		return Token{Type: EOF, Value: ""}
	}
	return p.tokens[p.current]
}

func (p *Parser) isAtEnd() bool {
	return p.current >= len(p.tokens) || p.tokens[p.current].Type == EOF
}
