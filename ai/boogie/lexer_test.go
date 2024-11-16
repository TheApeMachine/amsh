package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLexer(t *testing.T) {
	Convey("Given a boogie program", t, func() {
		program := `
		out <= () <= in
		`

		Convey("And a lexer", func() {
			lexer := NewLexer(program)

			Convey("When lexemes are generated", func() {

				count := 0
				ordered := []Lexeme{
					{Type: OUT, Value: "out"},
					{Type: OUTFLOW, Value: "<="},
					{Type: DELIMITER, Value: "("},
					{Type: DELIMITER, Value: ")"},
					{Type: OUTFLOW, Value: "<="},
					{Type: IN, Value: "in"},
				}

				for lexeme := range lexer.Generate() {
					So(lexeme, ShouldResemble, ordered[count])
					count++
				}
			})
		})
	})
}
