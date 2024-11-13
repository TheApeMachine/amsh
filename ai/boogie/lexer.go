package boogie

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/theapemachine/amsh/errnie"
)

// TokenType represents the type of token.
type TokenType int

const (
	NONE TokenType = iota
	EOF
	IDENTIFIER
	NUMBER
	BEHAVIOR
	PARAMETER
	ARROW
	DELIMITER
	OPERATOR
	OUTCOME
	COMMENT
	STRING
	LABEL
	JUMP
	SWITCH
	SELECT
	MATCH
	JOIN
)

type Token struct {
	Type  TokenType
	Value string
}

func (tokenType TokenType) String() string {
	return [...]string{
		"NONE",
		"EOF",
		"IDENTIFIER",
		"NUMBER",
		"BEHAVIOR",
		"PARAMETER",
		"ARROW",
		"DELIMITER",
		"OPERATOR",
		"OUTCOME",
		"COMMENT",
		"STRING",
		"LABEL",
		"JUMP",
		"SWITCH",
		"SELECT",
		"MATCH",
		"JOIN",
	}[tokenType]
}

// Lexer represents a state machine for lexing boogie code.
type Lexer struct {
	input     string
	position  int        // Current position in the input
	buffer    strings.Builder
	state     TokenType
	lastToken Token      // Last token processed
}

// NewLexer creates a new lexer instance
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:     input,
		position:  0,
		buffer:    strings.Builder{},
		state:     NONE,
		lastToken: Token{Type: NONE, Value: ""},
	}
}

// NextToken returns the next token from the input.
func (lexer *Lexer) Generate() chan Token {
	out := make(chan Token, 256)

	go func() {
		defer close(out)

		reader := bufio.NewReader(strings.NewReader(lexer.input))
		var (
			err  error
			char rune
		)

		for {
			if char, _, err = reader.ReadRune(); errnie.Error(err) != nil {
				if err == io.EOF {
					break
				}

				return
			}

			if !unicode.IsSpace(char) {
				lexer.buffer.WriteRune(char)
			}

			switch lexer.state {
			case NONE:
				lexer.state = lexer.nextState(char)
			case STRING:
				lexer.state = lexer.nextState(char)
				if lexer.state == NONE {
					lexer.sendAndReset(out)
				}
			}
		}
	}()

	return out
}

func (lexer *Lexer) sendAndReset(out chan Token) {
	out <- Token{Type: lexer.state, Value: lexer.buffer.String()}
	lexer.buffer.Reset()
}

func (lexer *Lexer) nextState(char rune) TokenType {
	switch {
	case char == ';':
		return COMMENT
	case char == '[':
		return PARAMETER
	case char == '=' && lexer.buffer.String() == "<":
		return ARROW
	case char == '=' && lexer.buffer.String() == ">":
		return ARROW
	case char == '|':
		return OPERATOR
	}

	switch lexer.buffer.String() {
	case "switch":
		return SWITCH
	case "select":
		return SELECT
	case "match":
		return MATCH
	case "join":
		return JOIN
	case "next", "back", "send", "cancel":
		return OUTCOME
	}

	if lexer.isDelimiter(char) {
		if lexer.isStringEnd(char) {
			return NONE
		}
		return DELIMITER
	}

	if unicode.IsLetter(char) {
		return IDENTIFIER
	}

	if unicode.IsDigit(char) {
		return NUMBER
	}

	return NONE
}

/*
isToken returns true if the current buffer definitely
contains a token.
*/
func (lexer *Lexer) isToken(state TokenType) bool {
	return !map[TokenType]bool{
		NONE:    true,
		STRING:  true,
		COMMENT: true,
	}[state]
}

func (lexer *Lexer) isDelimiter(char rune) bool {
	return map[rune]bool{
		'(': true,
		')': true,
		'<': true,
		'>': true,
		'"': true,
		'[': true,
		']': true,
		';': true,
		'|': true,
		'.': true,
	}[char]
}

func (lexer *Lexer) isStringEnd(char rune) bool {
	return lexer.state == STRING && char == '"'
}

// NextToken returns the next token in the input stream
func (lexer *Lexer) NextToken() Token {
	token := lexer.nextTokenInternal()
	lexer.lastToken = token
	return token
}

func (lexer *Lexer) nextTokenInternal() Token {
	lexer.skipWhitespace()

	if lexer.position >= len(lexer.input) {
		return Token{Type: EOF, Value: ""}
	}

	char := rune(lexer.input[lexer.position])

	switch {
	case char == ';':
		return lexer.readComment()
	case char == '<':
		if lexer.isArrow() {
			return lexer.readArrow()
		}
		return lexer.readBehavior()
	case char == '=':
		if lexer.isArrow() {
			return lexer.readArrow()
		}
	case char == '[':
		return lexer.readBracket()
	case char == '|':
		lexer.position++
		return Token{Type: OPERATOR, Value: "|"}
	case lexer.isDelimiter(char):
		lexer.position++
		return Token{Type: DELIMITER, Value: string(char)}
	}

	return lexer.readIdentifier()
}

func (lexer *Lexer) readComment() Token {
	startPos := lexer.position
	for lexer.position < len(lexer.input) && lexer.input[lexer.position] != '\n' {
		lexer.position++
	}
	return Token{Type: COMMENT, Value: string(lexer.input[startPos:lexer.position])}
}

func (lexer *Lexer) readArrow() Token {
	startPos := lexer.position
	// Read either <= or =>
	if lexer.position+1 < len(lexer.input) {
		if lexer.input[lexer.position:lexer.position+2] == "<=" ||
		   lexer.input[lexer.position:lexer.position+2] == "=>" {
			lexer.position += 2
			return Token{Type: ARROW, Value: string(lexer.input[startPos:lexer.position])}
		}
	}
	// If not an arrow, advance one position and let the next iteration handle it
	lexer.position++
	return Token{Type: DELIMITER, Value: string(lexer.input[startPos:lexer.position])}
}

func (lexer *Lexer) readBehavior() Token {
	startPos := lexer.position
	// Read until closing >
	for lexer.position < len(lexer.input) && lexer.input[lexer.position] != '>' {
		lexer.position++
	}
	if lexer.position < len(lexer.input) {
		lexer.position++ // include the closing >
	}
	return Token{Type: BEHAVIOR, Value: string(lexer.input[startPos:lexer.position])}
}

func (lexer *Lexer) readBracket() Token {
	startPos := lexer.position
	// Read until closing ]
	for lexer.position < len(lexer.input) && lexer.input[lexer.position] != ']' {
		lexer.position++
	}
	if lexer.position < len(lexer.input) {
		lexer.position++ // include the closing ]
	}
	
	value := string(lexer.input[startPos:lexer.position])
	// Determine if this is a jump, label, or parameter
	if strings.Contains(value, ".jump") {
		lexer.lastToken = Token{Type: JUMP, Value: value}
		return lexer.lastToken
	} 
	
	if lexer.lastToken.Type == SWITCH || lexer.lastToken.Type == SELECT {
		lexer.lastToken = Token{Type: LABEL, Value: value}
		return lexer.lastToken
	}
	
	lexer.lastToken = Token{Type: PARAMETER, Value: value}
	return lexer.lastToken
}

func (lexer *Lexer) readIdentifier() Token {
	startPos := lexer.position
	// Read until whitespace or delimiter
	for lexer.position < len(lexer.input) && 
		!lexer.isDelimiter(rune(lexer.input[lexer.position])) && 
		!unicode.IsSpace(rune(lexer.input[lexer.position])) {
		lexer.position++
	}
	
	value := string(lexer.input[startPos:lexer.position])
	// Check for keywords
	switch value {
	case "switch":
		return Token{Type: SWITCH, Value: value}
	case "select":
		return Token{Type: SELECT, Value: value}
	case "match":
		return Token{Type: MATCH, Value: value}
	case "join":
		return Token{Type: JOIN, Value: value}
	case "next", "back", "send", "cancel":
		return Token{Type: OUTCOME, Value: value}
	default:
		return Token{Type: IDENTIFIER, Value: value}
	}
}

func (lexer *Lexer) skipWhitespace() {
	for lexer.position < len(lexer.input) && 
		unicode.IsSpace(rune(lexer.input[lexer.position])) {
		lexer.position++
	}
}

func (lexer *Lexer) isArrow() bool {
	if lexer.position+1 >= len(lexer.input) {
		return false
	}
	return lexer.input[lexer.position:lexer.position+2] == "<=" ||
		   lexer.input[lexer.position:lexer.position+2] == "=>"
}
