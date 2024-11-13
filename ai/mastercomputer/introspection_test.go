// introspection_test.go
package mastercomputer

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIntrospectionSystem(t *testing.T) {
	Convey("Given an introspection system", t, func() {
		// Create system without real toolset for testing
		introspection := NewIntrospectionSystem(nil)

		// Mock some tools for testing
		mockTool := Capability{
			Name:        "mock_tool",
			Type:        "tool",
			Description: "A mock tool for testing",
			Inputs: map[string]string{
				"input1": "string",
				"input2": "int",
			},
			Outputs: map[string]string{
				"output1": "string",
			},
		}
		introspection.MockCapability(mockTool)

		Convey("When querying process capabilities", func() {
			processes := introspection.ListProcesses()

			Convey("It should return core processes", func() {
				So(len(processes), ShouldBeGreaterThan, 0)

				// Check surface process
				surface, exists := introspection.GetCapability("surface")
				So(exists, ShouldBeTrue)
				So(surface.Type, ShouldEqual, "process")
				So(surface.Behaviors, ShouldContain, "temporal")

				// Check quantum process
				quantum, exists := introspection.GetCapability("quantum")
				So(exists, ShouldBeTrue)
				So(quantum.Type, ShouldEqual, "process")
				So(quantum.Behaviors, ShouldContain, "superposition")
			})

			Convey("Process capabilities should include necessary information", func() {
				for _, proc := range processes {
					So(proc.Name, ShouldNotBeEmpty)
					So(proc.Description, ShouldNotBeEmpty)
					So(proc.Behaviors, ShouldNotBeEmpty)
					So(proc.Examples, ShouldNotBeEmpty)
				}
			})
		})

		Convey("When querying tool capabilities", func() {
			tools := introspection.ListTools()

			Convey("It should return registered tools", func() {
				So(len(tools), ShouldBeGreaterThan, 0)
				mockTool, exists := introspection.GetCapability("mock_tool")
				So(exists, ShouldBeTrue)
				So(mockTool.Type, ShouldEqual, "tool")
			})

			Convey("Tool capabilities should include necessary information", func() {
				So(mockTool.Name, ShouldEqual, "mock_tool")
				So(mockTool.Type, ShouldEqual, "tool")
				So(mockTool.Inputs, ShouldContainKey, "input1")
			})
		})

		Convey("When validating behaviors", func() {
			Convey("It should validate correct behaviors", func() {
				So(introspection.ValidateBehavior("surface", "temporal"), ShouldBeTrue)
				So(introspection.ValidateBehavior("quantum", "superposition"), ShouldBeTrue)
			})

			Convey("It should reject invalid behaviors", func() {
				So(introspection.ValidateBehavior("surface", "invalid"), ShouldBeFalse)
				So(introspection.ValidateBehavior("invalid", "temporal"), ShouldBeFalse)
			})
		})

		Convey("When querying context possibilities", func() {
			ctx := ProgramContext{
				CurrentProcess: "surface",
				PreviousSteps:  []string{"quantum"},
				State: map[string]interface{}{
					"status": "processing",
				},
			}

			possibilities := introspection.ContextPossibilities(ctx)

			Convey("It should return valid next steps", func() {
				So(possibilities, ShouldContainKey, "next_steps")
				So(possibilities, ShouldContainKey, "valid_behaviors")
				So(possibilities, ShouldContainKey, "error_handlers")
			})
		})
	})
}
