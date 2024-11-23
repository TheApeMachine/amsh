package marvin

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/amsh/ai/provider"
)

func TestBuffer(t *testing.T) {
	Convey("Given a new Buffer", t, func() {
		buffer := NewBuffer()

		Convey("When created", func() {
			Convey("It should have empty messages", func() {
				So(len(buffer.messages), ShouldEqual, 0)
			})

			Convey("It should have default max context tokens", func() {
				So(buffer.maxContextTokens, ShouldEqual, 10000)
			})
		})

		Convey("When adding messages with Poke", func() {
			msg1 := provider.Message{Role: "system", Content: "Test system message"}
			msg2 := provider.Message{Role: "user", Content: "Test user message"}

			buffer.Poke(msg1).Poke(msg2)

			Convey("It should contain the messages in order", func() {
				So(len(buffer.messages), ShouldEqual, 2)
				So(buffer.messages[0].Role, ShouldEqual, "system")
				So(buffer.messages[0].Content, ShouldEqual, "Test system message")
				So(buffer.messages[1].Role, ShouldEqual, "user")
				So(buffer.messages[1].Content, ShouldEqual, "Test user message")
			})
		})

		Convey("When peeking at messages", func() {
			msg := provider.Message{Role: "system", Content: "Test message"}
			buffer.Poke(msg)

			messages := buffer.Peek()

			Convey("It should return all messages", func() {
				So(len(messages), ShouldEqual, 1)
				So(messages[0].Role, ShouldEqual, "system")
				So(messages[0].Content, ShouldEqual, "Test message")
			})
		})

		Convey("When clearing the buffer", func() {
			buffer.Poke(provider.Message{Role: "system", Content: "Test message"})
			buffer.Clear()

			Convey("It should remove all messages", func() {
				So(len(buffer.messages), ShouldEqual, 0)
			})
		})

		Convey("When truncating messages", func() {
			// Add system and user messages (always kept)
			buffer.Poke(provider.Message{Role: "system", Content: "System prompt"})
			buffer.Poke(provider.Message{Role: "user", Content: "User message"})

			// Add several assistant messages
			for i := 0; i < 5; i++ {
				buffer.Poke(provider.Message{
					Role:    "assistant",
					Content: "This is a long message that should consume tokens " + "very long content ",
				})
			}

			truncated := buffer.Truncate()

			Convey("It should always keep system and user messages", func() {
				So(len(truncated), ShouldBeGreaterThanOrEqualTo, 2)
				So(truncated[0].Role, ShouldEqual, "system")
				So(truncated[1].Role, ShouldEqual, "user")
			})

			Convey("It should respect max token limit", func() {
				totalTokens := 0
				for _, msg := range truncated {
					totalTokens += buffer.estimateTokens(msg)
				}
				So(totalTokens, ShouldBeLessThanOrEqualTo, buffer.maxContextTokens)
			})
		})

		Convey("When converting buffer to string", func() {
			buffer.Poke(provider.Message{Role: "system", Content: "Test message"})
			str := buffer.String()

			Convey("It should format messages correctly", func() {
				So(str, ShouldContainSubstring, "<buffer>")
				So(str, ShouldContainSubstring, "<system>")
				So(str, ShouldContainSubstring, "Test message")
				So(str, ShouldContainSubstring, "</system>")
				So(str, ShouldContainSubstring, "</buffer>")
			})
		})
	})
}
