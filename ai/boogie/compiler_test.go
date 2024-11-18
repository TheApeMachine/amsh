package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCompiler(t *testing.T) {
	Convey("Given a parsed Boogie program", t, func() {
		Convey("When compiling an empty program", func() {
			program := `out <= () <= in`
			lexer := NewLexer()
			parser := NewParser()
			ast := parser.Generate(lexer.Generate(program))

			compiler := NewCompiler()
			compiler.Generate(ast)

			Convey("It should generate no instructions", func() {
				So(len(compiler.instructions), ShouldEqual, 0)
			})
		})

		Convey("When compiling a single operation program", func() {
			program := `out <= (analyze => send) <= in`
			lexer := NewLexer()
			parser := NewParser()
			ast := parser.Generate(lexer.Generate(program))

			compiler := NewCompiler()
			compiler.Generate(ast)

			Convey("It should generate two instructions", func() {
				So(len(compiler.instructions), ShouldEqual, 2)

				Convey("The first instruction should be analyze", func() {
					So(compiler.instructions[0].Type, ShouldEqual, INSTRUCTION_SPAWN)
					So(compiler.instructions[0].Operation, ShouldEqual, "analyze")
				})

				Convey("The second instruction should be send", func() {
					So(compiler.instructions[1].Type, ShouldEqual, INSTRUCTION_SPAWN)
					So(compiler.instructions[1].Operation, ShouldEqual, "send")
				})
			})
		})

		Convey("When compiling a multi operation program", func() {
			program := `out <= (
				analyze => next
				verify  => send
			) <= in`
			lexer := NewLexer()
			parser := NewParser()
			ast := parser.Generate(lexer.Generate(program))

			compiler := NewCompiler()
			compiler.Generate(ast)

			Convey("It should generate four instructions", func() {
				So(len(compiler.instructions), ShouldEqual, 4)

				Convey("Instructions should be in correct order with proper next pointers", func() {
					So(compiler.instructions[0].Operation, ShouldEqual, "analyze")
					So(compiler.instructions[1].Operation, ShouldEqual, "next")
					So(compiler.instructions[2].Operation, ShouldEqual, "verify")
					So(compiler.instructions[3].Operation, ShouldEqual, "send")
				})
			})
		})

		Convey("When compiling a program with parallel execution", func() {
			program := `
			out <= (
				join <= (
					analyze => next
					verify  => send
				) (
					analyze => next
					verify  => send
				)
			) <= in`
			lexer := NewLexer()
			parser := NewParser()
			ast := parser.Generate(lexer.Generate(program))

			compiler := NewCompiler()
			compiler.Generate(ast)

			Convey("It should generate nine instructions", func() {
				So(len(compiler.instructions), ShouldEqual, 9)

				Convey("The last instruction should be a join", func() {
					joinInstruction := compiler.instructions[8]
					So(joinInstruction.Type, ShouldEqual, INSTRUCTION_JOIN)
					So(len(joinInstruction.Next), ShouldEqual, 2)
				})

				Convey("First parallel branch should have correct operations", func() {
					So(compiler.instructions[0].Operation, ShouldEqual, "analyze")
					So(compiler.instructions[1].Operation, ShouldEqual, "next")
					So(compiler.instructions[2].Operation, ShouldEqual, "verify")
					So(compiler.instructions[3].Operation, ShouldEqual, "send")
				})

				Convey("Second parallel branch should have correct operations", func() {
					So(compiler.instructions[4].Operation, ShouldEqual, "analyze")
					So(compiler.instructions[5].Operation, ShouldEqual, "next")
					So(compiler.instructions[6].Operation, ShouldEqual, "verify")
					So(compiler.instructions[7].Operation, ShouldEqual, "send")
				})
			})
		})
	})
}
