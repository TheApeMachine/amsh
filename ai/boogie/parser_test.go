package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const whenParsed = "When parsed"
const shouldBeProgram = "It should be a program"
const shouldHaveOneChild = "It should have one child"
const shouldHaveNoChildren = "It should have no children"
const shouldHaveTwoChildren = "It should have two children"
const shouldHaveFourChildren = "It should have four children"
const shouldHaveClosure = "It should have a closure"
const shouldHaveNextOp = "It should have a next operation"
const shouldHaveJoin = "It should have a join"
const shouldHaveAnalysis = "It should have an analysis operation"
const shouldHaveVerify = "It should have a verify operation"
const shouldHaveSend = "It should have a send operation"

/*
TestParser checks that the parser correctly builds ASTs from Boogie programs.
*/
func TestParser(t *testing.T) {
	Convey("Given an empty Boogie program", t, func() {
		program := `out <= () <= in`
		lexer := NewLexer()
		parser := NewParser()

		Convey(whenParsed, func() {
			ast := parser.Generate(lexer.Generate(program))

			Convey(shouldBeProgram, func() {
				So(ast.Type, ShouldEqual, NODE_PROGRAM)
			})

			Convey(shouldHaveOneChild, func() {
				So(len(ast.Next), ShouldEqual, 1)
			})

			closure := ast.Next[0]

			Convey(shouldHaveClosure, func() {
				So(closure.Type, ShouldEqual, NODE_CLOSURE)
			})

			Convey(shouldHaveNoChildren, func() {
				So(len(closure.Next), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a single operation Boogie program", t, func() {
		program := `out <= (analyze => send) <= in`
		lexer := NewLexer()
		parser := NewParser()

		Convey(whenParsed, func() {
			ast := parser.Generate(lexer.Generate(program))

			Convey(shouldBeProgram, func() {
				So(ast.Type, ShouldEqual, NODE_PROGRAM)
			})

			Convey(shouldHaveOneChild, func() {
				So(len(ast.Next), ShouldEqual, 1)

				closure := ast.Next[0]

				Convey(shouldHaveClosure, func() {
					So(closure.Type, ShouldEqual, NODE_CLOSURE)
				})

				Convey(shouldHaveTwoChildren, func() {
					So(len(closure.Next), ShouldEqual, 2)
				})

				analyze := closure.Next[0]

				Convey(shouldHaveAnalysis, func() {
					So(analyze.Type, ShouldEqual, NODE_OPERATION)
					So(analyze.Value, ShouldEqual, "analyze")
				})

				send := closure.Next[1]

				Convey(shouldHaveSend, func() {
					So(send.Type, ShouldEqual, NODE_OPERATION)
					So(send.Value, ShouldEqual, "send")
				})
			})
		})
	})

	Convey("Given a multi operation Boogie program", t, func() {
		program := `out <= (
			analyze => next
			verify  => send
		) <= in`
		lexer := NewLexer()
		parser := NewParser()

		Convey(whenParsed, func() {
			ast := parser.Generate(lexer.Generate(program))

			Convey(shouldBeProgram, func() {
				So(ast.Type, ShouldEqual, NODE_PROGRAM)
			})

			Convey(shouldHaveOneChild, func() {
				So(len(ast.Next), ShouldEqual, 1)
			})

			closure := ast.Next[0]

			Convey(shouldHaveClosure, func() {
				So(closure.Type, ShouldEqual, NODE_CLOSURE)
			})

			Convey(shouldHaveFourChildren, func() {
				So(len(closure.Next), ShouldEqual, 4)
			})
			analyze := closure.Next[0]

			Convey(shouldHaveAnalysis, func() {
				So(analyze.Type, ShouldEqual, NODE_OPERATION)
				So(analyze.Value, ShouldEqual, "analyze")
			})

			next := closure.Next[1]

			Convey(shouldHaveNextOp, func() {
				So(next.Type, ShouldEqual, NODE_OPERATION)
				So(next.Value, ShouldEqual, "next")
			})

			verify := closure.Next[2]

			Convey(shouldHaveVerify, func() {
				So(verify.Type, ShouldEqual, NODE_OPERATION)
				So(verify.Value, ShouldEqual, "verify")
			})

			send := closure.Next[3]

			Convey(shouldHaveSend, func() {
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
		lexer := NewLexer()
		parser := NewParser()

		Convey(whenParsed, func() {
			ast := parser.Generate(lexer.Generate(program))

			Convey("It should create the correct AST structure", func() {
				Convey("It should be a program", func() {
					So(ast.Type, ShouldEqual, NODE_PROGRAM)
				})

				Convey(shouldHaveOneChild, func() {
					So(len(ast.Next), ShouldEqual, 1)
				})

				closure := ast.Next[0]

				Convey(shouldHaveClosure, func() {
					So(closure.Type, ShouldEqual, NODE_CLOSURE)
				})

				Convey(shouldHaveJoin, func() {
					So(closure.Next[0].Type, ShouldEqual, NODE_JOIN)
				})

				join := closure.Next[0]

				Convey(shouldHaveTwoChildren, func() {
					So(len(join.Next), ShouldEqual, 2)

					Convey("It should have a closure for the first child", func() {
						So(join.Next[0].Type, ShouldEqual, NODE_CLOSURE)

						Convey(shouldHaveFourChildren, func() {
							So(len(join.Next[0].Next), ShouldEqual, 4)

							Convey(shouldHaveAnalysis, func() {
								So(join.Next[0].Next[0].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[0].Next[0].Value, ShouldEqual, "analyze")
							})

							Convey(shouldHaveNextOp, func() {
								So(join.Next[0].Next[1].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[0].Next[1].Value, ShouldEqual, "next")
							})

							Convey(shouldHaveVerify, func() {
								So(join.Next[0].Next[2].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[0].Next[2].Value, ShouldEqual, "verify")
							})

							Convey(shouldHaveSend, func() {
								So(join.Next[0].Next[3].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[0].Next[3].Value, ShouldEqual, "send")
							})
						})
					})

					Convey("It should have a closure for the second child", func() {
						So(join.Next[1].Type, ShouldEqual, NODE_CLOSURE)

						Convey(shouldHaveFourChildren, func() {
							So(len(join.Next[1].Next), ShouldEqual, 4)

							Convey(shouldHaveAnalysis, func() {
								So(join.Next[1].Next[0].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[1].Next[0].Value, ShouldEqual, "analyze")
							})

							Convey(shouldHaveNextOp, func() {
								So(join.Next[1].Next[1].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[1].Next[1].Value, ShouldEqual, "next")
							})

							Convey(shouldHaveVerify, func() {
								So(join.Next[1].Next[2].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[1].Next[2].Value, ShouldEqual, "verify")
							})

							Convey(shouldHaveSend, func() {
								So(join.Next[1].Next[3].Type, ShouldEqual, NODE_OPERATION)
								So(join.Next[1].Next[3].Value, ShouldEqual, "send")
							})
						})
					})
				})

			})
		})
	})
}
