package provider

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAccumulator(t *testing.T) {
	Convey("Given an Accumulator", t, func() {
		accumulator := NewAccumulator()

		Convey("When streaming events", func() {
			input := make(chan Event)
			output := make(chan Event, 10)

			go accumulator.Stream(input, output)

			Convey("Should handle multiple events from same agent", func() {
				events := []Event{
					{AgentID: "agent1", Content: "Hello"},
					{AgentID: "agent1", Content: " World"},
				}

				for _, event := range events {
					input <- event
				}
				close(input)

				// Allow time for processing
				time.Sleep(10 * time.Millisecond)

				Convey("Should accumulate events correctly", func() {
					result := accumulator.String()
					So(result, ShouldEqual, "Hello World")
				})

				Convey("Should forward events to sink", func() {
					var received []Event
					for i := 0; i < 2; i++ {
						select {
						case event := <-output:
							received = append(received, event)
						case <-time.After(10 * time.Millisecond):
							// timeout
						}
					}
					So(len(received), ShouldEqual, 2)
					So(received[0].Content, ShouldEqual, "Hello")
					So(received[1].Content, ShouldEqual, " World")
				})
			})
		})

		Convey("When collecting events", func() {
			input := make(chan Event)

			go func() {
				input <- Event{AgentID: "agent1", Content: "Hello"}
				input <- Event{AgentID: "agent1", Content: " "}
				input <- Event{AgentID: "agent1", Content: "World"}
				close(input)
			}()

			result := accumulator.Collect(input)

			Convey("Should concatenate event content in order", func() {
				So(result, ShouldEqual, "Hello World")
			})
		})

		Convey("When handling multiple agents", func() {
			input := make(chan Event)
			output1 := make(chan Event, 10)
			output2 := make(chan Event, 10)

			go accumulator.Stream(input, output1, output2)

			events := []Event{
				{AgentID: "agent1", Content: "Hello"},
				{AgentID: "agent2", Content: "Hi"},
				{AgentID: "agent1", Content: " World"},
				{AgentID: "agent2", Content: " There"},
			}

			for _, event := range events {
				input <- event
			}
			close(input)

			time.Sleep(10 * time.Millisecond)

			Convey("Should maintain separate buffers for each agent", func() {
				result := accumulator.String()
				So(result, ShouldContainSubstring, "Hello World")
				So(result, ShouldContainSubstring, "Hi There")
			})

			Convey("Should forward to multiple sinks", func() {
				for i := 0; i < 2; i++ {
					event1 := <-output1
					event2 := <-output2
					So(event1, ShouldResemble, event2)
				}
			})
		})

		Convey("When sink channel is full", func() {
			input := make(chan Event)
			fullSink := make(chan Event) // unbuffered channel that's never read

			go accumulator.Stream(input, fullSink)

			Convey("Should not block on full sink", func() {
				done := make(chan bool)
				go func() {
					input <- Event{AgentID: "agent1", Content: "Test"}
					close(input)
					done <- true
				}()

				select {
				case <-done:
					// Success - didn't block
				case <-time.After(100 * time.Millisecond):
					t.Error("Stream blocked on full sink")
				}
			})
		})
	})
}
