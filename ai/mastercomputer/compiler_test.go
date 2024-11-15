package mastercomputer

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/amsh/ai/boogie"
)

func TestCompiler(t *testing.T) {
	convey.Convey("Given a new Compiler instance", t, func() {
		compiler := NewCompiler()

		convey.Convey("When compiling a basic program", func() {
			program := &boogie.Program{
				Input:  "input_channel",
				Output: "output_channel",
				Root: &boogie.Operation{
					Type:     "basic",
					Behavior: "test",
					Parameters: []string{
						"param1", "value1",
						"param2", "value2",
					},
				},
			}

			instructions := compiler.Compile(program)

			convey.Convey("Then it should generate the correct instruction sequence", func() {
				convey.So(len(instructions), convey.ShouldEqual, 3)
				convey.So(instructions[0].Op, convey.ShouldEqual, OpReceive)
				convey.So(instructions[1].Op, convey.ShouldEqual, OpCall)
				convey.So(instructions[2].Op, convey.ShouldEqual, OpSend)
			})
		})

		convey.Convey("When compiling a control flow program", func() {
			program := &boogie.Program{
				Input:  "input_channel",
				Output: "output_channel",
				Root: &boogie.Operation{
					Type: "switch",
					Parameters: []string{
						"condition", "true",
					},
					Children: []*boogie.Operation{
						{
							Type:     "basic",
							Behavior: "path1",
						},
						{
							Type:     "basic",
							Behavior: "path2",
						},
					},
				},
			}

			instructions := compiler.Compile(program)

			convey.Convey("Then it should generate branch instructions", func() {
				convey.So(len(instructions), convey.ShouldBeGreaterThan, 3)
				convey.So(instructions[1].Op, convey.ShouldEqual, OpBranch)
			})
		})

		convey.Convey("When compiling a match operation", func() {
			program := &boogie.Program{
				Input:  "input_channel",
				Output: "output_channel",
				Root: &boogie.Operation{
					Type: "match",
					Children: []*boogie.Operation{
						{
							Type:       "basic",
							Parameters: []string{"pattern", "case1"},
						},
						{
							Type:       "basic",
							Parameters: []string{"pattern", "case2"},
						},
					},
				},
			}

			instructions := compiler.Compile(program)

			convey.Convey("Then it should generate match instructions", func() {
				convey.So(len(instructions), convey.ShouldBeGreaterThan, 3)
				foundBranch := false
				for _, inst := range instructions {
					if inst.Op == OpBranch {
						foundBranch = true
						break
					}
				}
				convey.So(foundBranch, convey.ShouldBeTrue)
			})
		})

		convey.Convey("When compiling a join operation", func() {
			program := &boogie.Program{
				Input:  "input_channel",
				Output: "output_channel",
				Root: &boogie.Operation{
					Type: "join",
					Children: []*boogie.Operation{
						{
							Type:     "basic",
							Behavior: "concurrent1",
						},
						{
							Type:     "basic",
							Behavior: "concurrent2",
						},
					},
				},
			}

			instructions := compiler.Compile(program)

			convey.Convey("Then it should generate join instructions", func() {
				convey.So(len(instructions), convey.ShouldBeGreaterThan, 3)
				foundJoin := false
				for _, inst := range instructions {
					if inst.Op == OpJoin {
						foundJoin = true
						convey.So(inst.Operands[0], convey.ShouldEqual, 2) // Number of concurrent operations
						break
					}
				}
				convey.So(foundJoin, convey.ShouldBeTrue)
			})
		})

		convey.Convey("When compiling a select operation", func() {
			program := &boogie.Program{
				Input:  "input_channel",
				Output: "output_channel",
				Root: &boogie.Operation{
					Type: "select",
					Parameters: []string{
						"condition", "loop",
					},
					Children: []*boogie.Operation{
						{
							Type:     "basic",
							Behavior: "loop_body",
						},
					},
				},
			}

			instructions := compiler.Compile(program)

			convey.Convey("Then it should generate loop instructions", func() {
				convey.So(len(instructions), convey.ShouldBeGreaterThan, 3)
				foundJump := false
				for _, inst := range instructions {
					if inst.Op == OpJump {
						foundJump = true
						break
					}
				}
				convey.So(foundJump, convey.ShouldBeTrue)
			})
		})
	})
}
