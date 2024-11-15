package boogie

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLexer(t *testing.T) {
	Convey("Given a boogie lexer", t, func() {
		Convey("When lexing a simple operation", func() {
			input := "analyze<temporal> => next | cancel"
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: IDENTIFIER, Value: "analyze"},
				{Type: BEHAVIOR, Value: "<temporal>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "cancel"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing a tool call with parameters", func() {
			input := "[query, \"data\"]call<browser> => send | back"
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: PARAMETER, Value: "[query, \"data\"]"},
				{Type: IDENTIFIER, Value: "call"},
				{Type: BEHAVIOR, Value: "<browser>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "send"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "back"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing a complete program", func() {
			input := `
				out <= (
					analyze<temporal> => next | cancel
					[search, "quantum"]call<browser> => send | back
				) <= in ; this is a comment
			`
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: IDENTIFIER, Value: "out"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: IDENTIFIER, Value: "analyze"},
				{Type: BEHAVIOR, Value: "<temporal>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "cancel"},
				{Type: PARAMETER, Value: "[search, \"quantum\"]"},
				{Type: IDENTIFIER, Value: "call"},
				{Type: BEHAVIOR, Value: "<browser>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "send"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "back"},
				{Type: DELIMITER, Value: ")"},
				{Type: ARROW, Value: "<="},
				{Type: IDENTIFIER, Value: "in"},
				{Type: COMMENT, Value: "; this is a comment"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing control flow keywords", func() {
			input := "switch <= ( select <= ( match <= ( join <= () ) ) )"
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: SWITCH, Value: "switch"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: SELECT, Value: "select"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: MATCH, Value: "match"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: JOIN, Value: "join"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: DELIMITER, Value: ")"},
				{Type: DELIMITER, Value: ")"},
				{Type: DELIMITER, Value: ")"},
				{Type: DELIMITER, Value: ")"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing labels and jumps", func() {
			input := "switch[mylabel] <= ( [mylabel.jump] )"
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: SWITCH, Value: "switch"},
				{Type: LABEL, Value: "[mylabel]"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: JUMP, Value: "[mylabel.jump]"},
				{Type: DELIMITER, Value: ")"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing behaviors with numbers", func() {
			input := "reason <= <5> => next"
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: IDENTIFIER, Value: "reason"},
				{Type: ARROW, Value: "<="},
				{Type: BEHAVIOR, Value: "<5>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing complex nested structures", func() {
			input := `
				switch <= (
					success => send
					default => [step2.analyze.jump]
				) <= match <= (
					<5> => send
					default => [mylabel.jump]
				)
			`
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: SWITCH, Value: "switch"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: IDENTIFIER, Value: "success"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "send"},
				{Type: IDENTIFIER, Value: "default"},
				{Type: ARROW, Value: "=>"},
				{Type: JUMP, Value: "[step2.analyze.jump]"},
				{Type: DELIMITER, Value: ")"},
				{Type: ARROW, Value: "<="},
				{Type: MATCH, Value: "match"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: BEHAVIOR, Value: "<5>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "send"},
				{Type: IDENTIFIER, Value: "default"},
				{Type: ARROW, Value: "=>"},
				{Type: JUMP, Value: "[mylabel.jump]"},
				{Type: DELIMITER, Value: ")"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing comments", func() {
			input := "analyze => next ; this is a comment\nvalidate => send"
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: IDENTIFIER, Value: "analyze"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: COMMENT, Value: "; this is a comment"},
				{Type: IDENTIFIER, Value: "validate"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "send"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing a pipeline with multiple outcomes", func() {
			input := "validate => next | back | cancel"
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: IDENTIFIER, Value: "validate"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "back"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "cancel"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})

		Convey("When lexing a complex program with multiple features", func() {
			input := `
				; This is a complex program that demonstrates various features
				out <= (
					switch[mainflow] <= (
						; First analyze the input
						[data, "complex query"]analyze<temporal> => next | back | cancel

						; Then process with browser
						success => (
							select <= (
								[browser.settings]call<browser> => send | back
								match <= (
									<5> => next    ; Try up to 5 times
									default => [mainflow.jump]
								)
							)
						)

						; Handle errors and join results
						error => join <= (
							[error.log]validate => next | cancel
							[results]process<parallel> => send | back
						)

						default => [error.handler.jump]
					)
				) <= in
			`
			lexer := NewLexer(input)
			tokens := collectTokens(lexer)

			expected := []Token{
				{Type: COMMENT, Value: "; This is a complex program that demonstrates various features"},
				{Type: IDENTIFIER, Value: "out"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: SWITCH, Value: "switch"},
				{Type: LABEL, Value: "[mainflow]"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: COMMENT, Value: "; First analyze the input"},
				{Type: PARAMETER, Value: "[data, \"complex query\"]"},
				{Type: IDENTIFIER, Value: "analyze"},
				{Type: BEHAVIOR, Value: "<temporal>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "back"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "cancel"},
				{Type: COMMENT, Value: "; Then process with browser"},
				{Type: IDENTIFIER, Value: "success"},
				{Type: ARROW, Value: "=>"},
				{Type: DELIMITER, Value: "("},
				{Type: SELECT, Value: "select"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: PARAMETER, Value: "[browser.settings]"},
				{Type: IDENTIFIER, Value: "call"},
				{Type: BEHAVIOR, Value: "<browser>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "send"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "back"},
				{Type: MATCH, Value: "match"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: BEHAVIOR, Value: "<5>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: COMMENT, Value: "; Try up to 5 times"},
				{Type: IDENTIFIER, Value: "default"},
				{Type: ARROW, Value: "=>"},
				{Type: JUMP, Value: "[mainflow.jump]"},
				{Type: DELIMITER, Value: ")"},
				{Type: DELIMITER, Value: ")"},
				{Type: DELIMITER, Value: ")"},
				{Type: COMMENT, Value: "; Handle errors and join results"},
				{Type: IDENTIFIER, Value: "error"},
				{Type: ARROW, Value: "=>"},
				{Type: JOIN, Value: "join"},
				{Type: ARROW, Value: "<="},
				{Type: DELIMITER, Value: "("},
				{Type: PARAMETER, Value: "[error.log]"},
				{Type: IDENTIFIER, Value: "validate"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "next"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "cancel"},
				{Type: PARAMETER, Value: "[results]"},
				{Type: IDENTIFIER, Value: "process"},
				{Type: BEHAVIOR, Value: "<parallel>"},
				{Type: ARROW, Value: "=>"},
				{Type: OUTCOME, Value: "send"},
				{Type: OPERATOR, Value: "|"},
				{Type: OUTCOME, Value: "back"},
				{Type: DELIMITER, Value: ")"},
				{Type: IDENTIFIER, Value: "default"},
				{Type: ARROW, Value: "=>"},
				{Type: JUMP, Value: "[error.handler.jump]"},
				{Type: DELIMITER, Value: ")"},
				{Type: DELIMITER, Value: ")"},
				{Type: ARROW, Value: "<="},
				{Type: IDENTIFIER, Value: "in"},
				{Type: EOF, Value: ""},
			}

			So(len(tokens), ShouldEqual, len(expected))
			for i, token := range tokens {
				So(token.Type, ShouldEqual, expected[i].Type)
				So(strings.TrimSpace(token.Value), ShouldEqual, strings.TrimSpace(expected[i].Value))
			}
		})
	})
}

func collectTokens(lexer *Lexer) []Token {
	tokens := []Token{}
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}
