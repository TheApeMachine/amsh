package ai

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuffer(t *testing.T) {
	Convey("Given a new Buffer", t, func() {
		systemPrompt := "You are a test agent."
		userPrompt := "Your task is to test things."
		buffer := NewBuffer(systemPrompt, userPrompt)

		Convey("When initialized", func() {
			Convey("Then it should have correct initial prompts", func() {
				messages := buffer.GetMessages()
				So(len(messages), ShouldEqual, 2)
				So(messages[0].Role, ShouldEqual, "system")
				So(messages[0].Content, ShouldEqual, systemPrompt)
				So(messages[1].Role, ShouldEqual, "user")
				So(messages[1].Content, ShouldEqual, userPrompt)
			})

			Convey("Then it should have empty message history", func() {
				So(len(buffer.messages), ShouldEqual, 0)
			})
		})

		Convey("When adding messages", func() {
			buffer.AddMessage("user", "Hello")
			buffer.AddMessage("assistant", "Hi there")

			Convey("Then messages should be added to history", func() {
				messages := buffer.GetMessages()
				So(len(messages), ShouldEqual, 4) // 2 prompts + 2 messages
				So(messages[2].Role, ShouldEqual, "user")
				So(messages[2].Content, ShouldEqual, "Hello")
				So(messages[3].Role, ShouldEqual, "assistant")
				So(messages[3].Content, ShouldEqual, "Hi there")
			})

			Convey("Then messages should maintain order", func() {
				messages := buffer.GetMessages()
				roles := []string{"system", "user", "user", "assistant"}
				for i, msg := range messages {
					So(msg.Role, ShouldEqual, roles[i])
				}
			})
		})

		Convey("When adding tool results", func() {
			buffer.AddToolResult("search", "Found something interesting")

			Convey("Then tool result should be formatted correctly", func() {
				messages := buffer.GetMessages()
				lastMsg := messages[len(messages)-1]
				So(lastMsg.Role, ShouldEqual, "tool")
				So(lastMsg.Content, ShouldEqual, "Tool search returned: Found something interesting")
			})
		})

		Convey("When clearing the buffer", func() {
			buffer.AddMessage("user", "Test message")
			buffer.AddMessage("assistant", "Test response")
			buffer.Clear()

			Convey("Then message history should be empty", func() {
				So(len(buffer.messages), ShouldEqual, 0)
			})

			Convey("Then prompts should still be preserved", func() {
				messages := buffer.GetMessages()
				So(len(messages), ShouldEqual, 2)
				So(messages[0].Role, ShouldEqual, "system")
				So(messages[1].Role, ShouldEqual, "user")
			})
		})

		Convey("When handling empty prompts", func() {
			emptyBuffer := NewBuffer("", "")

			Convey("Then it should handle empty system prompt", func() {
				messages := emptyBuffer.GetMessages()
				So(len(messages), ShouldEqual, 0)
			})

			Convey("Then it should still accept new messages", func() {
				emptyBuffer.AddMessage("user", "Test")
				messages := emptyBuffer.GetMessages()
				So(len(messages), ShouldEqual, 1)
				So(messages[0].Role, ShouldEqual, "user")
			})
		})

		Convey("When handling message capacity", func() {
			// Add many messages
			for i := 0; i < 100; i++ {
				buffer.AddMessage("user", "Test message")
			}

			Convey("Then it should handle large message counts", func() {
				messages := buffer.GetMessages()
				So(len(messages), ShouldEqual, 102) // 2 prompts + 100 messages
			})

			Convey("Then it should maintain correct order with many messages", func() {
				messages := buffer.GetMessages()
				So(messages[0].Role, ShouldEqual, "system")
				So(messages[1].Role, ShouldEqual, "user")
				So(messages[2].Role, ShouldEqual, "user")
			})
		})

		Convey("When handling concurrent operations", func() {
			done := make(chan bool)
			for i := 0; i < 10; i++ {
				go func() {
					buffer.AddMessage("user", "Concurrent message")
					buffer.GetMessages()
					done <- true
				}()
			}

			Convey("Then it should handle concurrent access safely", func() {
				// Wait for all goroutines to complete
				for i := 0; i < 10; i++ {
					<-done
				}
				messages := buffer.GetMessages()
				So(len(messages), ShouldBeGreaterThan, 2)
			})
		})
	})
}
