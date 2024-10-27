package ai

import (
	"bytes"
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/provider"
	"github.com/theapemachine/amsh/ai/types"
)

func TestTeam(t *testing.T) {
	Convey("Given a new Team", t, func() {
		// Setup test configuration
		viper.SetConfigType("yaml")
		testConfig := []byte(`
ai:
    prompt:
        system: |
            You are {name}, {role}.
            {job_description}
        researcher:
            role: |
                You are a researcher focused on gathering information.
        analyst:
            role: |
                You are an analyst focused on processing information.
        processes:
            discussion: |
                Team discussion process.
toolsets:
    base:
        - memory
    researcher:
        - memory
        - browser
    analyst:
        - memory
tools:
    memory:
        description: |
            Access and manipulate the memory storage systems.
        parameters:
            store:
                type: string
                enum: [vector, graph]
                description: The type of storage to use.
            operation:
                type: string
                enum: [store, retrieve, search, delete]
                description: The operation to perform.
            data:
                type: string
                description: The data to store or query parameters.
`)
		viper.ReadConfig(bytes.NewBuffer(testConfig))

		toolset := NewTestToolset()
		team := NewTeam(toolset)
		So(team, ShouldNotBeNil)

		Convey("When initializing agents", func() {
			researcher := team.GetResearcher()
			analyst := team.GetAnalyst()

			Convey("Then it should create agents with correct roles", func() {
				So(researcher, ShouldNotBeNil)
				So(analyst, ShouldNotBeNil)
				So(researcher.GetRole(), ShouldEqual, types.RoleResearcher)
				So(analyst.GetRole(), ShouldEqual, types.RoleAnalyst)
			})

			Convey("Then agents should start in idle state", func() {
				So(researcher.GetState(), ShouldEqual, types.StateIdle)
				So(analyst.GetState(), ShouldEqual, types.StateIdle)
			})

			Convey("Then agents should have appropriate tools", func() {
				researcherTools := researcher.GetTools()
				analystTools := analyst.GetTools()
				So(researcherTools, ShouldNotBeNil)
				So(analystTools, ShouldNotBeNil)
			})
		})

		Convey("When setting the provider", func() {
			mockProvider := &MockProvider{}
			team.SetProvider(mockProvider)

			Convey("Then all agents should receive the provider", func() {
				researcher := team.GetResearcher()
				analyst := team.GetAnalyst()
				So(researcher.provider, ShouldEqual, mockProvider)
				So(analyst.provider, ShouldEqual, mockProvider)
			})
		})

		Convey("When getting an agent by role", func() {
			Convey("Then it should return the correct agent for valid roles", func() {
				researcher := team.GetAgent("researcher")
				So(researcher, ShouldNotBeNil)
				So(researcher.GetRole(), ShouldEqual, types.RoleResearcher)
			})

			Convey("Then it should return nil for invalid roles", func() {
				invalidAgent := team.GetAgent("invalid_role")
				So(invalidAgent, ShouldBeNil)
			})
		})

		Convey("When shutting down the team", func() {
			team.SetProvider(&MockProvider{})
			researcher := team.GetResearcher()
			analyst := team.GetAnalyst()

			// Set initial states
			researcher.SetState(types.StateWorking)
			analyst.SetState(types.StateWorking)

			team.Shutdown()

			Convey("Then all agents should be in done state", func() {
				So(researcher.GetState(), ShouldEqual, types.StateDone)
				So(analyst.GetState(), ShouldEqual, types.StateDone)
			})

			Convey("Then all agents should have cleared their tools", func() {
				So(researcher.GetTools(), ShouldBeEmpty)
				So(analyst.GetTools(), ShouldBeEmpty)
			})
		})

		Convey("When handling concurrent operations", func() {
			Convey("Then it should safely handle multiple goroutines", func() {
				done := make(chan bool)
				for i := 0; i < 10; i++ {
					go func() {
						team.GetResearcher()
						team.GetAnalyst()
						team.SetProvider(&MockProvider{})
						done <- true
					}()
				}

				// Wait for all goroutines to complete
				for i := 0; i < 10; i++ {
					<-done
				}
				// If we reach here without panic, the test passes
				So(true, ShouldBeTrue)
			})
		})
	})
}

func (m *MockProvider) GenerateStream(ctx context.Context, messages []provider.Message) (<-chan string, <-chan error) {
	responseChan := make(chan string)
	errChan := make(chan error)
	go func() {
		responseChan <- "mock stream response"
		close(responseChan)
		close(errChan)
	}()
	return responseChan, errChan
}
