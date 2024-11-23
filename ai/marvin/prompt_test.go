package marvin

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
)

func TestPrompt(t *testing.T) {
	// Setup test configuration
	viper.SetDefault("ai.setups.marvin.templates.system", "System template")
	viper.SetDefault("ai.setups.marvin.templates.test_role", "Test role template")

	Convey("Given a new Prompt", t, func() {
		prompt := NewPrompt("test_role")

		Convey("When created with a role", func() {
			Convey("It should load correct templates", func() {
				So(prompt.systemPrompt, ShouldEqual, "System template")
				So(prompt.rolePrompt, ShouldEqual, "Test role template")
			})

			Convey("It should initialize with empty user prompt", func() {
				So(prompt.userPrompt, ShouldBeEmpty)
			})

			Convey("It should initialize with empty processes", func() {
				So(len(prompt.processes), ShouldEqual, 0)
			})
		})

		Convey("When setting user prompt", func() {
			testPrompt := "Test user prompt"
			prompt.SetUserPrompt(testPrompt)

			Convey("It should store the user prompt", func() {
				So(prompt.userPrompt, ShouldEqual, testPrompt)
			})
		})

		Convey("When getting System message", func() {
			// Add a test process
			testProcess := &MockProcess{schema: "test schema"}
			prompt.AddProcess(testProcess)

			message := prompt.System()

			Convey("It should return correct message structure", func() {
				So(message.Role, ShouldEqual, "system")
				So(message.Content, ShouldContainSubstring, "System template")
				So(message.Content, ShouldContainSubstring, "Test role template")
				So(message.Content, ShouldContainSubstring, "<schema>")
				So(message.Content, ShouldContainSubstring, "test schema")
				So(message.Content, ShouldContainSubstring, "</schema>")
			})
		})

		Convey("When getting User message", func() {
			testPrompt := "Test user input"
			prompt.SetUserPrompt(testPrompt)
			message := prompt.User()

			Convey("It should return correct message structure", func() {
				So(message.Role, ShouldEqual, "user")
				So(message.Content, ShouldEqual, testPrompt)
			})
		})

		Convey("When getting Context message", func() {
			testPrompt := "Test context"
			prompt.SetUserPrompt(testPrompt)
			message := prompt.Context()

			Convey("It should return correct message structure", func() {
				So(message.Role, ShouldEqual, "assistant")
				So(message.Content, ShouldContainSubstring, "<context>")
				So(message.Content, ShouldContainSubstring, testPrompt)
				So(message.Content, ShouldContainSubstring, "</context>")
				So(message.Content, ShouldContainSubstring, "Please respond according to the schema provided")
			})
		})

		Convey("When adding processes", func() {
			process1 := &MockProcess{schema: "schema1"}
			process2 := &MockProcess{schema: "schema2"}

			prompt.AddProcess(process1)
			prompt.AddProcess(process2)

			Convey("It should store all processes", func() {
				So(len(prompt.processes), ShouldEqual, 2)
			})

			Convey("Processes should be reflected in System message", func() {
				message := prompt.System()
				So(message.Content, ShouldContainSubstring, "schema1")
				So(message.Content, ShouldContainSubstring, "schema2")
			})
		})
	})
}

// MockProcess implements the Process interface for testing
type MockProcess struct {
	schema string
}

func (mp *MockProcess) GenerateSchema() string {
	return mp.schema
}
