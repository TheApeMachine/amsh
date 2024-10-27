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

func (m *MockProvider) GenerateSync(ctx context.Context, messages []provider.Message) (string, error) {
	return m.response, nil
}

// Add the Generate method to satisfy the Provider interface
func (m *MockProvider) Generate(ctx context.Context, messages []provider.Message) <-chan provider.Event {
	responseChan := make(chan provider.Event, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errChan)
		responseChan <- provider.Event{
			Type:    provider.EventToken,
			Content: m.response,
			Error:   nil,
		}
	}()

	return responseChan
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
			So(agent.context, ShouldEqual, systemPrompt)
			So(agent.task, ShouldEqual, userPrompt)
		})

		Convey("When executing a task", func() {
			// Clear the buffer before executing the task
			agent.buffer = NewBuffer(systemPrompt, userPrompt)

			response, err := agent.ExecuteTask()

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)
				So(response, ShouldEqual, "Test response")
				So(agent.GetState(), ShouldEqual, types.StateDone)
				So(agent.buffer.GetMessages(), ShouldHaveLength, 3) // Only expecting the assistant's response
			})

			Convey("And the buffer should contain the conversation", func() {
				messages := agent.buffer.GetMessages()
				So(len(messages), ShouldEqual, 3) // Only expecting the assistant's response
				So(messages[2].Role, ShouldEqual, "assistant")
				So(messages[2].Content, ShouldEqual, "Test response")
				So(agent.buffer.GetMessages(), ShouldHaveLength, 3) // Only expecting the assistant's response
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
				So(agent.tools, ShouldBeNil)
			})
		})
	})
}
