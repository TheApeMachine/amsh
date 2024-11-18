package boogie

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/smartystreets/goconvey/convey"
)

/*
TestParser checks that the parser correctly builds ASTs from Boogie programs.
*/
func TestParser(t *testing.T) {
	Convey("Given an empty Boogie program", t, func() {
		program := `out <= () <= in`
		lexer := NewLexer(program)
		parser := NewParser(lexer.Generate())

		Convey("When parsed", func() {
			ast := parser.Generate()

			Convey("It should create the correct AST structure", func() {
				So(ast.Type, ShouldEqual, NODE_PROGRAM)
				So(len(ast.Next), ShouldEqual, 1)

				closure := ast.Next[0]
				So(closure.Type, ShouldEqual, NODE_CLOSURE)
				So(len(closure.Next), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a single operation Boogie program", t, func() {
		program := `out <= (analyze => send) <= in`
		lexer := NewLexer(program)
		parser := NewParser(lexer.Generate())

		Convey("When parsed", func() {
			ast := parser.Generate()

			Convey("It should create the correct AST structure", func() {
				So(ast.Type, ShouldEqual, NODE_PROGRAM)
				So(len(ast.Next), ShouldEqual, 1)

				closure := ast.Next[0]
				So(closure.Type, ShouldEqual, NODE_CLOSURE)
				So(len(closure.Next), ShouldEqual, 2)

				analyze := closure.Next[0]
				So(analyze.Type, ShouldEqual, NODE_OPERATION)
				So(analyze.Value, ShouldEqual, "analyze")

				send := closure.Next[1]
				So(send.Type, ShouldEqual, NODE_OPERATION)
				So(send.Value, ShouldEqual, "send")
			})
		})
	})

	Convey("Given a multi operation Boogie program", t, func() {
		program := `out <= (
			analyze => next
			verify  => send
		) <= in`
		lexer := NewLexer(program)
		parser := NewParser(lexer.Generate())

		Convey("When parsed", func() {
			ast := parser.Generate()

			Convey("It should create the correct AST structure", func() {
				So(ast.Type, ShouldEqual, NODE_PROGRAM)
				So(len(ast.Next), ShouldEqual, 1)

				closure := ast.Next[0]
				So(closure.Type, ShouldEqual, NODE_CLOSURE)
				So(len(closure.Next), ShouldEqual, 4)

				analyze := closure.Next[0]
				So(analyze.Type, ShouldEqual, NODE_OPERATION)
				So(analyze.Value, ShouldEqual, "analyze")

				next := closure.Next[1]
				So(next.Type, ShouldEqual, NODE_OPERATION)
				So(next.Value, ShouldEqual, "next")

				verify := closure.Next[2]
				So(verify.Type, ShouldEqual, NODE_OPERATION)
				So(verify.Value, ShouldEqual, "verify")

				send := closure.Next[3]
				So(send.Type, ShouldEqual, NODE_OPERATION)
				So(send.Value, ShouldEqual, "send")
			})
		})
	})

	Convey("Given a program with parallel execution", t, func() {
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
		lexer := NewLexer(program)
		parser := NewParser(lexer.Generate())

		Convey("When parsed", func() {
			ast := parser.Generate()
			spew.Dump(ast)

			Convey("It should create the correct AST structure", func() {
				So(ast.Type, ShouldEqual, NODE_PROGRAM)
				So(len(ast.Next), ShouldEqual, 1)

				outerClosure := ast.Next[0]
				So(outerClosure.Type, ShouldEqual, NODE_CLOSURE)
				So(len(outerClosure.Next), ShouldEqual, 1)

				join := outerClosure.Next[0]
				So(join.Type, ShouldEqual, NODE_JOIN)
				So(len(join.Next), ShouldEqual, 2)

				// Check first closure
				closure1 := join.Next[0]
				So(closure1.Type, ShouldEqual, NODE_CLOSURE)
				So(len(closure1.Next), ShouldEqual, 4)

				// Check second closure
				closure2 := join.Next[1]
				So(closure2.Type, ShouldEqual, NODE_CLOSURE)
				So(len(closure2.Next), ShouldEqual, 4)
			})
		})
	})
}
