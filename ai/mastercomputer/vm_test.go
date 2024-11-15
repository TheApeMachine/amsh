// vm_test.go
package mastercomputer

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	ShouldExecuteSuccessfully = "should execute successfully"
)

func TestAgentVM(t *testing.T) {
	Convey("Given an agent virtual machine", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		comm := NewAgentCommunication(ctx)
		vm := NewAgentVM(comm.pool, comm)

		Reset(func() {
			cancel()
			comm.Close()
		})

		Convey("When executing basic memory operations", func() {
			instructions := []Instruction{
				{Op: OpStore, Operands: []interface{}{"test-key", "test-value"}},
				{Op: OpLoad, Operands: []interface{}{"test-key"}},
			}

			err := vm.Execute(ctx, instructions)

			Convey(ShouldExecuteSuccessfully, func() {
				So(err, ShouldBeNil)
				So(vm.state.Accumulator, ShouldEqual, "test-value")
			})
		})

		Convey("When executing communication operations", func() {
			// Set up test agents
			agent1 := NewAgent(ctx, "test")
			agent2 := NewAgent(ctx, "test")
			comm.RegisterAgent(agent1)
			comm.RegisterAgent(agent2)

			instructions := []Instruction{
				{Op: OpStore, Operands: []interface{}{"message", "Hello!"}},
				{Op: OpLoad, Operands: []interface{}{"message"}},
				{Op: OpSend, Operands: []interface{}{"agent2", "Hello!"}},
			}

			err := vm.Execute(ctx, instructions)

			Convey(ShouldExecuteSuccessfully, func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When executing control flow operations", func() {
			instructions := []Instruction{
				{Op: OpStore, Operands: []interface{}{"counter", 0}},
				{Op: OpLoad, Operands: []interface{}{"counter"}},
				{Op: OpBranch, Operands: []interface{}{true, 4}},
				{Op: OpNop},
				{Op: OpStore, Operands: []interface{}{"result", "branched"}},
			}

			err := vm.Execute(ctx, instructions)

			Convey(ShouldExecuteSuccessfully, func() {
				So(err, ShouldBeNil)
				So(vm.state.Memory["result"], ShouldEqual, "branched")
			})
		})

		Convey("When handling errors", func() {
			Convey("Invalid instruction should return error", func() {
				instructions := []Instruction{
					{Op: OpStore}, // Missing operands
				}

				err := vm.Execute(ctx, instructions)
				So(err, ShouldNotBeNil)
			})

			Convey("Invalid jump target should return error", func() {
				instructions := []Instruction{
					{Op: OpJump, Operands: []interface{}{999}}, // Out of bounds
				}

				err := vm.Execute(ctx, instructions)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "out of bounds")
			})
		})
	})
}
