package boogie

import (
	"strings"
	"unicode"
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

// Token represents a lexical token.
type Token struct {
	Type  TokenType
	Value string
}

// Lexer represents a state machine for lexing boogie code.
type Lexer struct {
	input     string
	position  int // Current position in the input
	buffer    strings.Builder
	state     TokenType
	lastToken Token // Last token processed
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input:     input,
		position:  0,
		buffer:    strings.Builder{},
		state:     NONE,
		lastToken: Token{Type: NONE, Value: ""},
	}
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
	case isDelimiter(char):
		lexer.position++
		return Token{Type: DELIMITER, Value: string(char)}
	}

	// Handle identifiers and keywords
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
	if lexer.position+1 < len(lexer.input) {
		possibleArrow := lexer.input[lexer.position : lexer.position+2]
		if possibleArrow == "<=" || possibleArrow == "=>" {
			lexer.position += 2
			return Token{Type: ARROW, Value: possibleArrow}
		}
	}
	lexer.position++
	return Token{Type: DELIMITER, Value: "<"}
}

func (lexer *Lexer) readBehavior() Token {
	startPos := lexer.position
	lexer.position++ // Skip '<'
	for lexer.position < len(lexer.input) && lexer.input[lexer.position] != '>' {
		lexer.position++
	}
	if lexer.position < len(lexer.input) && lexer.input[lexer.position] == '>' {
		lexer.position++ // Include '>'
	}
	return Token{Type: BEHAVIOR, Value: string(lexer.input[startPos:lexer.position])}
}

func (lexer *Lexer) readBracket() Token {
	startPos := lexer.position
	lexer.position++ // Skip '['
	for lexer.position < len(lexer.input) && lexer.input[lexer.position] != ']' {
		lexer.position++
	}
	if lexer.position < len(lexer.input) && lexer.input[lexer.position] == ']' {
		lexer.position++ // Include ']'
	}

	value := string(lexer.input[startPos:lexer.position])
	// Determine if this is a jump, label, or parameter
	if strings.Contains(value, ".jump") {
		return Token{Type: JUMP, Value: value}
	} else if lexer.lastToken.Type == SWITCH || lexer.lastToken.Type == SELECT {
		return Token{Type: LABEL, Value: value}
	}
	return Token{Type: PARAMETER, Value: value}
}

func (lexer *Lexer) readIdentifier() Token {
	startPos := lexer.position
	for lexer.position < len(lexer.input) &&
		!isDelimiter(rune(lexer.input[lexer.position])) &&
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

func isDelimiter(char rune) bool {
	return char == '(' || char == ')' ||
		char == '<' || char == '>' ||
		char == '"' || char == '[' ||
		char == ']' || char == ';' ||
		char == '|' || char == '.' || char == ','
}

// Helper function to parse parameters
func parseParameters(paramStr string) []string {
	raw := paramStr[1 : len(paramStr)-1] // Remove '[' and ']'
	parts := strings.Split(raw, ",")
	params := []string{}
	for _, part := range parts {
		params = append(params, strings.TrimSpace(part))
	}
	return params
}
