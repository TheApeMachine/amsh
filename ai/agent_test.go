package ai

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
)

type MockProvider struct {
	response string
}

// Update Generate to match the provider.Provider interface
func (m *MockProvider) Generate(ctx context.Context, messages []provider.Message) <-chan provider.Event {
	ch := make(chan provider.Event)
	go func() {
		defer close(ch)
		ch <- provider.Event{
			Content: m.response,
		}
	}()
	return ch
}

// Update GenerateSync to match the provider.Provider interface
func (m *MockProvider) GenerateSync(ctx context.Context, messages []provider.Message) (string, error) {
	return m.response, nil
}

func TestAgent(t *testing.T) {
	Convey("Given a new agent", t, func() {
		mockProvider := &MockProvider{response: "Test response"}

		// Create a minimal toolset for testing
		toolset := &Toolset{
			tools: map[string]types.Tool{
				"memory": types.NewBaseTool(
					"memory",
					"Test memory tool",
					nil, // parameters map
					func(ctx context.Context, args map[string]interface{}) (string, error) {
						return "test result", nil
					},
				),
			},
		}

		systemPrompt := "You are a riddle solver"
		userPrompt := "Solve this riddle: test"

		agent := NewAgent(
			"test-agent",
			types.RoleResearcher,
			systemPrompt,
			userPrompt,
			toolset,
			mockProvider,
		)

		Convey("When checking initial state", func() {
			So(agent.GetID(), ShouldEqual, "test-agent")
			So(agent.GetRole(), ShouldEqual, types.RoleResearcher)
			So(agent.GetState(), ShouldEqual, types.StateIdle)
			So(agent.Context, ShouldEqual, systemPrompt)
			So(agent.Task, ShouldEqual, userPrompt)
		})

		Convey("When executing a task", func() {
			// Clear the buffer before executing the task
			agent.Buffer = NewBuffer(systemPrompt, userPrompt)

			response, err := agent.ExecuteTask()

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)
				So(response, ShouldEqual, "Test response")
				So(agent.GetState(), ShouldEqual, types.StateDone)
				So(agent.Buffer.GetMessages(), ShouldHaveLength, 3)
			})

			Convey("And the buffer should contain the conversation", func() {
				messages := agent.Buffer.GetMessages()
				So(len(messages), ShouldEqual, 3)
				So(messages[2].Role, ShouldEqual, "assistant")
				So(messages[2].Content, ShouldEqual, "Test response")
				So(agent.Buffer.GetMessages(), ShouldHaveLength, 3)
			})
		})

		Convey("When receiving a message", func() {
			err := agent.ReceiveMessage("Test message")

			Convey("Then it should be added to the queue", func() {
				So(err, ShouldBeNil)
				So(agent.GetMessageCount(), ShouldEqual, 1)
			})
		})

		Convey("When shutting down", func() {
			agent.Shutdown()

			Convey("Then it should be in done state", func() {
				So(agent.GetState(), ShouldEqual, types.StateDone)
				So(agent.Tools, ShouldBeNil)
			})
		})
	})
}
