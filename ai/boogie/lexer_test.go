package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const whenTokenized = "When the program is tokenzized"

/*
TestLexer checks that the lexer correctly tokenizes Boogie programs.
*/
func TestLexerEmptyProgram(t *testing.T) {
	Convey("Given an empty Boogie program ", t, func() {
		program := `out <= () <= in`

		Convey(whenTokenized, func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer()

			for lexeme := range lexer.Generate(program) {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})
}

func TestLexerSingleOperation(t *testing.T) {
	Convey("Given a single operation Boogie program ", t, func() {
		program := `out <= (analyze => send) <= in`

		Convey(whenTokenized, func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer()

			for lexeme := range lexer.Generate(program) {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})
}

func TestLexerMultiOperation(t *testing.T) {
	Convey("Given a multi operation Boogie program ", t, func() {
		program := `out <= (
		    analyze => next
			verify  => send
		) <= in`

		Convey(whenTokenized, func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer()

			for lexeme := range lexer.Generate(program) {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "next"},
				{ID: OPERATION, Text: "verify"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})

	Convey("Given a multi operation Boogie program with fallbacks, behaviors and comments", t, func() {
		program := `
		out <= (
			analyze => next | back<3> | cancel ; Analysis with error handling
			verify  => send | back<3> | cancel ; Verification, and send up the chain
		) <= in
		`
		Convey(whenTokenized, func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer()

			for lexeme := range lexer.Generate(program) {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "next"},
				{ID: DELIMITER, Text: "|"},
				{ID: OPERATION, Text: "back"},
				{ID: DELIMITER, Text: "<"},
				{ID: VALUE, Text: "3"},
				{ID: DELIMITER, Text: ">"},
				{ID: DELIMITER, Text: "|"},
				{ID: OPERATION, Text: "cancel"},
				{ID: OPERATION, Text: "verify"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: DELIMITER, Text: "|"},
				{ID: OPERATION, Text: "back"},
				{ID: DELIMITER, Text: "<"},
				{ID: VALUE, Text: "3"},
				{ID: DELIMITER, Text: ">"},
				{ID: DELIMITER, Text: "|"},
				{ID: OPERATION, Text: "cancel"},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})
}

func TestLexerMultiOperationWithConditionalFlow(t *testing.T) {
	Convey("Given a multi operation Boogie program with conditional flow", t, func() {
		program := `
		out <= (
			analyze => next ; Analysis without any behavior
			verify  => next ; Verification, and send up the chain
			match (
				ok    => send   ; If ok, send up the chain
				error => cancel ; If error, cancel
			)
		) <= in
		`
		Convey(whenTokenized, func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer()

			for lexeme := range lexer.Generate(program) {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "next"},
				{ID: OPERATION, Text: "verify"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "next"},
				{ID: OPERATION, Text: "match"},
				{ID: DELIMITER, Text: "("},
				{ID: VALUE, Text: "ok"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: VALUE, Text: "error"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "cancel"},
				{ID: DELIMITER, Text: ")"},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})
}

func TestLexerMultiOperationWithParallelExecution(t *testing.T) {
	Convey("Given a multi operation Boogie program with parallel execution", t, func() {
		program := `
		out <= (
			join <= (
				analyze => next ; Analysis without any behavior
				verify  => send ; Verification, and send up the chain
			) (
				analyze => next ; Analysis without any behavior
				verify  => send ; Verification, and send up the chain
			)
		) <= in
		`
		Convey(whenTokenized, func() {
			lexemes := make([]Lexeme, 0)
			lexer := NewLexer()

			for lexeme := range lexer.Generate(program) {
				lexemes = append(lexemes, lexeme)
			}

			So(lexemes, ShouldResemble, []Lexeme{
				{ID: VALUE, Text: "out"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "join"},
				{ID: FLOW, Text: "<="},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "next"},
				{ID: OPERATION, Text: "verify"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: DELIMITER, Text: ")"},
				{ID: DELIMITER, Text: "("},
				{ID: OPERATION, Text: "analyze"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "next"},
				{ID: OPERATION, Text: "verify"},
				{ID: FLOW, Text: "=>"},
				{ID: OPERATION, Text: "send"},
				{ID: DELIMITER, Text: ")"},
				{ID: DELIMITER, Text: ")"},
				{ID: FLOW, Text: "<="},
				{ID: VALUE, Text: "in"},
			})
		})
	})
}
