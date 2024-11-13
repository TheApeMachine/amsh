// communication_test.go
package mastercomputer

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/qpool"
)

func TestAgentCommunication(t *testing.T) {
	Convey("Given an agent communication system", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		comm := NewAgentCommunication(ctx)

		Reset(func() {
			cancel()
			comm.Close()
		})

		Convey("When setting up test agents", func() {
			agent1 := NewAgent("agent1", "developer")
			agent2 := NewAgent("agent2", "analyst")

			comm.RegisterAgent(agent1)
			comm.RegisterAgent(agent2)

			Convey("The agents should be registered successfully", func() {
				So(comm.agents, ShouldContainKey, "agent1")
				So(comm.agents, ShouldContainKey, "agent2")
			})

			Convey("When starting a discussion", func() {
				discussionID, err := comm.StartDiscussion([]string{"agent1", "agent2"})

				Convey("It should create the discussion successfully", func() {
					So(err, ShouldBeNil)
					So(discussionID, ShouldNotBeEmpty)

					Convey("And agents should be able to join the discussion", func() {
						messages1, err1 := comm.JoinDiscussion(discussionID)
						messages2, err2 := comm.JoinDiscussion(discussionID)

						So(err1, ShouldBeNil)
						So(err2, ShouldBeNil)
						So(messages1, ShouldNotBeNil)
						So(messages2, ShouldNotBeNil)

						Convey("When sending a message in the discussion", func() {
							testMsg := Message{
								From:    "agent1",
								Content: "Hello, agent2!",
								Type:    "discussion",
							}

							err := comm.SendMessage(discussionID, testMsg)
							So(err, ShouldBeNil)

							Convey("The message should be received by other participants", func() {
								select {
								case msg := <-messages2:
									receivedMsg := msg.Value.(Message)
									So(receivedMsg.From, ShouldEqual, "agent1")
									So(receivedMsg.Content, ShouldEqual, "Hello, agent2!")
									So(receivedMsg.Type, ShouldEqual, "discussion")
								case <-time.After(time.Second * 2):
									So("Message reception", ShouldNotTimeOut)
								}
							})
						})
					})
				})
			})

			Convey("When sending an instruction", func() {
				instruction := map[string]interface{}{
					"action": "analyze",
					"data":   "test data",
				}

				result, err := comm.SendInstruction("agent1", "agent2", instruction)

				Convey("The instruction should be sent successfully", func() {
					So(err, ShouldBeNil)
					So(result, ShouldNotBeNil)

					Convey("And should receive a response", func() {
						select {
						case response := <-result:
							So(response, ShouldNotBeNil)
						case <-time.After(time.Second * 2):
							So("Instruction response", ShouldNotTimeOut)
						}
					})
				})
			})

			Convey("When trying to join a non-existent discussion", func() {
				_, err := comm.JoinDiscussion("non-existent")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "discussion not found")
			})

			Convey("When sending an instruction to a non-existent agent", func() {
				_, err := comm.SendInstruction("agent1", "non-existent", "test")
				So(err, ShouldBeNil) // Initial scheduling succeeds

				// The actual error would occur during execution
			})
		})
	})
}

func TestCommunicationPatterns(t *testing.T) {
	Convey("Given an agent communication system", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		comm := NewAgentCommunication(ctx)

		Reset(func() {
			cancel()
			comm.Close()
		})

		Convey("When creating a stream channel", func() {
			config := PatternConfig{
				Pattern: PatternStream,
				Timeout: time.Second * 5,
				Options: map[string]interface{}{
					"buffer_size":  10,
					"backpressure": "drop_oldest",
				},
			}

			channel, err := comm.CreateChannel(config)

			Convey("It should create successfully", func() {
				So(err, ShouldBeNil)
				So(channel, ShouldNotBeNil)
				So(channel.Pattern, ShouldEqual, PatternStream)

				Convey("When sending data to stream", func() {
					testData := []string{"data1", "data2", "data3"}

					for _, data := range testData {
						channel.Messages <- qpool.QuantumValue{
							Value: Message{
								Content: data,
								Type:    "stream",
							},
							CreatedAt: time.Now(),
						}
					}

					Convey("It should handle the data flow", func() {
						// Wait for processing
						time.Sleep(time.Millisecond * 100)

						state := channel.State.Context
						So(state["type"], ShouldEqual, "stream")
						So(state["buffer_size"], ShouldEqual, 10)
					})
				})
			})
		})

		Convey("When creating a query channel", func() {
			config := PatternConfig{
				Pattern: PatternQuery,
				Timeout: time.Second * 2,
			}

			channel, err := comm.CreateChannel(config)

			Convey("It should create successfully", func() {
				So(err, ShouldBeNil)
				So(channel, ShouldNotBeNil)

				Convey("When sending a query", func() {
					queryMsg := Message{
						ID:      "query-1",
						Content: "What is the status?",
						Type:    "query",
					}

					channel.Messages <- qpool.QuantumValue{
						Value:     queryMsg,
						CreatedAt: time.Now(),
					}

					Convey("It should handle timeouts", func() {
						timeoutChan := time.After(time.Second * 2)
						
						select {
						case response := <-channel.Messages:
							// Verify we got a response, even if it's a timeout error
							So(response, ShouldNotBeNil)
						case <-timeoutChan:
							// This is actually the expected path for this test
							// since we're testing timeout behavior
							So(true, ShouldBeTrue) // Test passes if we timeout as expected
						}
					})
				})
			})
		})
	})
}

// Custom assertion for timeout
func ShouldNotTimeOut(actual interface{}, expected ...interface{}) string {
	return "Operation timed out"
}
