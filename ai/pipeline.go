package ai

import (
	"context"

	"github.com/google/uuid"
	"github.com/theapemachine/amsh/ai/tools"
	"github.com/theapemachine/amsh/errnie"
)

/*
Chunk represents a chunk of data with its unique session ID, sequence ID, iteration, prompt, team, agent, and response.
It provides methods to initialize and generate chunks, as well as to retrieve and set various attributes.
*/
type Chunk struct {
	SessionID  string `json:"session_id"`
	SequenceID string `json:"sequence_id"`
	Iteration  int    `json:"iteration"`
	Team       *Team  `json:"team"`
	Agent      *Agent `json:"agent"`
	Response   string `json:"response"`
}

/*
Pipeline represents a pipeline of AI teams with a context and a list of teams.
It provides methods to initialize the pipeline, generate responses, and manage the current iteration.
*/
type Pipeline struct {
	ctx              context.Context
	sessionID        string
	teams            []*Team
	currentIteration int
}

/*
NewPipeline creates a new Pipeline with the provided context and connection to an AI service.
It initializes the pipeline with an empty list of teams and sets the current iteration to 0.
*/
func NewPipeline(ctx context.Context) *Pipeline {
	errnie.Trace()

	return &Pipeline{
		ctx:              ctx,
		sessionID:        uuid.New().String(),
		teams:            make([]*Team, 0),
		currentIteration: 0,
	}
}

/*
Initialize is a method that initializes the pipeline by adding teams to the pipeline.
It creates teams with the specified names and agents, starts them, and adds them to the pipeline.
*/
func (pipeline *Pipeline) Initialize() *Pipeline {
	errnie.Trace()

	for _, teamName := range []string{
		"ingress",
	} {
		team := NewTeam(
			pipeline.ctx,
			teamName,
			NewAgent(
				pipeline.ctx,
				"prompt_engineer",
				[]tools.Tool{},
			),
			NewAgent(
				pipeline.ctx,
				"prompt_engineer",
				[]tools.Tool{},
			),
			NewAgent(
				pipeline.ctx,
				"prompt_engineer",
				[]tools.Tool{},
			),
		)

		team.Start() // Ensure this is called
		pipeline.teams = append(pipeline.teams, team)
	}

	for _, team := range pipeline.teams {
		team.Initialize()

		for _, agent := range team.Agents {
			agent.Prompt.System = append(agent.Prompt.System, team.Prompt.System...)
		}
	}

	return pipeline
}

/*
Generate is a method that generates responses based on a given prompt and number of iterations.
It creates a new chunk with the provided prompt and iterates through the teams to generate responses.
The responses are sent to the output channel, which is returned to the caller.
*/
func (pipeline *Pipeline) Generate(prompt string, iterations int) <-chan Chunk {
	errnie.Trace()
	out := make(chan Chunk)

	go func() {
		defer close(out)

		for _, team := range pipeline.teams {
			team.SetPrompt(prompt)

			chunk := Chunk{
				SessionID:  pipeline.sessionID,
				SequenceID: uuid.New().String(),
				Iteration:  pipeline.currentIteration,
				Team:       team,
			}

			for chunkOut := range team.Generate(chunk) {
				out <- chunkOut
			}
		}
	}()

	return out
}
