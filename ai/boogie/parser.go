package boogie

import (
	"errors"
	"strings"
)

/*
AST represents the abstract syntax tree of a Boogie program.
*/
type AST struct {
	Operations []Operation
}

/*
Operation represents a single operation in a Boogie program.
*/
type Operation struct {
	Source      string
	Destination string
	Action      string
}

/*
Parse converts a sequence of tokens into an abstract syntax tree (AST).
*/
func Parse(tokens []string) (*AST, error) {
	p := &parser{tokens: tokens, position: 0}
	return p.parseProgram()
}

type parser struct {
	tokens   []string
	position int
}

func (p *parser) parseProgram() (*AST, error) {
	if !p.match("out", "<=", "(") {
		return nil, errors.New("expected program to start with 'out <= ('")
	}

	operations, err := p.parseOperations()
	if err != nil {
		return nil, err
	}

	if !p.match(")", "<=", "in") {
		return nil, errors.New("expected program to end with ') <= in'")
	}

	if operations == nil {
		operations = []Operation{}
	}

	return &AST{Operations: operations}, nil
}

func (p *parser) parseOperations() ([]Operation, error) {
	var operations []Operation

	for !p.check(")") && !p.isAtEnd() {
		source, err := p.consume()
		if err != nil {
			return nil, err
		}

		// Handle label declarations
		if strings.HasPrefix(source, "[") && strings.HasSuffix(source, "]") {
			if !p.match("=>") {
				return nil, errors.New("expected '=>' after label declaration")
			}

			if !p.check("(") {
				return nil, errors.New("expected '(' after label declaration and '=>'")
			}
			p.advance() // Consume the '('
			
			nestedOps, err := p.parseOperations()
			if err != nil {
				return nil, err
			}

			// Add the label operation
			operations = append(operations, Operation{
				Source:      source,
				Action:      "=>",
				Destination: "analyze", // Connect label to first operation
			})
			
			// Add all nested operations
			operations = append(operations, nestedOps...)
			
			if !p.match(")") {
				return nil, errors.New("expected closing ')' after labeled block")
			}
			continue
		}

		// Handle match construct
		if source == "match" {
			operations = append(operations, Operation{
				Source:      "match",
				Destination: "ok",
				Action:      "=>",
			})
			if !p.match("(") {
				return nil, errors.New("expected '(' after 'match'")
			}
			for !p.check(")") && !p.isAtEnd() {
				condition, err := p.consume()
				if err != nil {
					return nil, err
				}
				action, err := p.consume()
				if err != nil {
					return nil, err
				}
				destination, err := p.consume()
				if err != nil {
					return nil, err
				}
				operations = append(operations, Operation{
					Source:      condition,
					Destination: destination,
					Action:      action,
				})
			}
			if !p.match(")") {
				return nil, errors.New("expected closing ')' for match construct")
			}
			continue
		}

		action, err := p.consume()
		if err != nil {
			return nil, err
		}

		if p.check("(") {
			p.advance() // Consume the '('
			nestedOps, err := p.parseOperations()
			if err != nil {
				return nil, err
			}
			operations = append(operations, nestedOps...)
			if !p.match(")") {
				return nil, errors.New("expected closing ')' for nested closure")
			}
		} else {
			destination, err := p.consume()
			if err != nil {
				return nil, err
			}

			operations = append(operations, Operation{
				Source:      source,
				Destination: destination,
				Action:      action,
			})
		}

		if p.check("|") {
			p.advance()
			destination, err := p.consume()
			if err != nil {
				return nil, err
			}
			operations = append(operations, Operation{
				Source:      source,
				Destination: destination,
				Action:      "|",
			})
		}
	}

	return operations, nil
}

func (p *parser) match(expected ...string) bool {
	for _, exp := range expected {
		if !p.check(exp) {
			return false
		}
		p.advance()
	}
	return true
}

func (p *parser) check(expected string) bool {
	if p.isAtEnd() {
		return false
	}
	return p.tokens[p.position] == expected
}

func (p *parser) advance() {
	if !p.isAtEnd() {
		p.position++
	}
}

func (p *parser) isAtEnd() bool {
	return p.position >= len(p.tokens)
}

func (p *parser) consume() (string, error) {
	if p.isAtEnd() {
		return "", errors.New("unexpected end of input")
	}
	token := p.tokens[p.position]
	p.advance()
	return token, nil
}
