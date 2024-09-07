package ai

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/sashabaranov/go-openai"
)

func TestNewAgent(t *testing.T) {
	Convey("Given a Conn, RoleType, and Tools", t, func() {
		conn := NewConn()
		role := CODER
		tools := []Tool{&mockTool{}}

		Convey("When creating a new Agent", func() {
			agent := NewAgent(conn, role, tools, "test")

			Convey("It should not be nil", func() {
				So(agent, ShouldNotBeNil)
			})

			Convey("It should have the correct Conn", func() {
				So(agent.conn, ShouldEqual, conn)
			})

			Convey("It should have the correct RoleType", func() {
				So(agent.role, ShouldEqual, role)
			})

			Convey("It should have the correct Tools", func() {
				So(agent.tools, ShouldResemble, tools)
			})
		})
	})
}

func TestGetToolDefinitions(t *testing.T) {
	Convey("Given an Agent with Tools", t, func() {
		agent := NewAgent(NewConn(), CODER, []Tool{&mockTool{}}, "test")

		Convey("When getting tool definitions", func() {
			definitions := agent.getToolDefinitions()

			Convey("It should return the correct number of definitions", func() {
				So(len(definitions), ShouldEqual, 1)
			})

			Convey("The definition should match the tool's definition", func() {
				So(definitions[0], ShouldResemble, mockToolDefinition)
			})
		})
	})
}

// Mock Tool for testing
type mockTool struct{}

func (m *mockTool) Read(p []byte) (n int, err error)  { return 0, nil }
func (m *mockTool) Write(p []byte) (n int, err error) { return 0, nil }
func (m *mockTool) Close() error                      { return nil }

var mockToolDefinition = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        "mock_tool",
		Description: "A mock tool for testing",
	},
}

func (m *mockTool) Definition() openai.Tool {
	return mockToolDefinition
}