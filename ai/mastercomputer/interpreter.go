package mastercomputer

import (
	"context"
	"errors"
	"strings"

	"github.com/theapemachine/amsh/ai/boogie"
	"github.com/theapemachine/amsh/ai/provider"
)

type Interpreter struct {
	currentToken int
	tokens       []string
}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (interpreter *Interpreter) Parse(pipeline string) (*boogie.Program, error) {
	// Tokenize the pipeline
	interpreter.tokens = interpreter.tokenize(pipeline)
	interpreter.currentToken = 0

	return interpreter.parseProgram()
}

func (i *Interpreter) tokenize(pipeline string) []string {
	// Basic tokenization - this would need to be more sophisticated
	pipeline = strings.ReplaceAll(pipeline, "<=", " <= ")
	pipeline = strings.ReplaceAll(pipeline, "=>", " => ")
	pipeline = strings.ReplaceAll(pipeline, "(", " ( ")
	pipeline = strings.ReplaceAll(pipeline, ")", " ) ")
	pipeline = strings.ReplaceAll(pipeline, "|", " | ")

	return strings.Fields(pipeline)
}

func (interpreter *Interpreter) parseProgram() (*boogie.Program, error) {
	prog := boogie.NewProgram()

	// Expected format: out <= type <= ( operations ) <= in
	if len(interpreter.tokens) < 7 {
		return nil, errors.New("invalid pipeline format")
	}

	if interpreter.tokens[0] != "out" || interpreter.tokens[len(interpreter.tokens)-1] != "in" {
		return nil, errors.New("pipeline must start with 'out' and end with 'in'")
	}

	// Parse program type
	prog.Type = interpreter.tokens[2]

	// Parse operations
	operations, err := interpreter.parseOperations()
	if err != nil {
		return nil, err
	}
	prog.Operations = operations

	return prog, nil
}

// // Parse converts a pipeline string into a Program structure
// func (interpreter *Interpreter) Parse(pipeline string) error {
// 	// Delegate to interpreter
// 	parsed, err := interpreter.Parse(pipeline)
// 	if err != nil {
// 		return err
// 	}
// 	*p = *parsed
// 	return nil
// }

func (interpreter *Interpreter) Execute(
	ctx context.Context, program *boogie.Program,
) <-chan provider.Event {
	processor := NewProcessor()
	return processor.Execute(ctx, program)
}

func (interpreter *Interpreter) parseOperations() ([]boogie.Operation, error) {
	var operations []boogie.Operation

	for interpreter.currentToken < len(interpreter.tokens) {
		token := interpreter.tokens[interpreter.currentToken]

		switch token {
		case ")":
			return operations, nil
		case ";":
			interpreter.currentToken++
			continue
		default:
			op, err := interpreter.parseOperation()
			if err != nil {
				return nil, err
			}
			operations = append(operations, op)
		}
	}

	return operations, nil
}

func (interpreter *Interpreter) parseOperation() (boogie.Operation, error) {
	op := boogie.Operation{}

	// Parse name and label
	token := interpreter.tokens[interpreter.currentToken]
	if strings.Contains(token, "[") {
		// Handle labeled operation
		parts := strings.Split(strings.Trim(token, "[]"), ".")
		op.Name = parts[0]
		op.Label = parts[1]
	} else {
		op.Name = token
	}

	// Parse outcomes
	interpreter.currentToken++ // Move to =>
	interpreter.currentToken++ // Move to first outcome

	for interpreter.currentToken < len(interpreter.tokens) && interpreter.tokens[interpreter.currentToken] != ")" {
		token := interpreter.tokens[interpreter.currentToken]
		if token == "|" {
			interpreter.currentToken++
			continue
		}
		op.Outcomes = append(op.Outcomes, token)
		interpreter.currentToken++
	}

	return op, nil
}
