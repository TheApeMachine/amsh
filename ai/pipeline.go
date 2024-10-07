package ai

import (
	"context"
	"os"

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
	teams            map[string]*Team
	currentIteration int
	Response         string
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
		teams:            make(map[string]*Team),
		currentIteration: 0,
	}
}

/*
Initialize is a method that initializes the pipeline by adding teams to the pipeline.
It creates teams with the specified names and agents, starts them, and adds them to the pipeline.
*/
func (pipeline *Pipeline) Initialize() *Pipeline {
	errnie.Trace()

	for _, team := range [][]string{
		{"prompt_engineers", "prompt_engineer"},
		{"routing", "router"},
		{"reasoners", "reasoner"},
		{"verifiers", "verifier"},
		{"researchers", "researcher"},
		{"analysts", "analyst"},
	} {
		pipeline.teams[team[0]] = NewTeam(
			pipeline.ctx,
			team[0],
			NewAgent(
				pipeline.ctx,
				team[1],
				[]tools.Tool{},
			),
			NewAgent(
				pipeline.ctx,
				team[1],
				[]tools.Tool{},
			),
			NewAgent(
				pipeline.ctx,
				team[1],
				[]tools.Tool{},
			),
		)

		pipeline.teams[team[0]].Initialize()
		pipeline.teams[team[0]].Start()
	}

	return pipeline
}

func (pipeline *Pipeline) newChunk() Chunk {
	errnie.Trace()

	return Chunk{
		SessionID:  pipeline.sessionID,
		SequenceID: uuid.New().String(),
		Iteration:  pipeline.currentIteration,
	}
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

		var (
			team     *Team
			chunk    Chunk
			response string
		)

		for {
			team = pipeline.teams["routing"]
			team.SetPrompt(prompt)
			chunk = pipeline.newChunk()
			chunk.Team = team

			response = ""
			for chunkOut := range team.Generate(chunk) {
				response += chunkOut.Response
				out <- chunkOut
			}

			command := NewAgent(pipeline.ctx, "command", []tools.Tool{})
			command.SetPrompt(response)
			command.Start()
			chunk = pipeline.newChunk()
			chunk.Team = &Team{
				ID: "command",
			}

			response = ""
			for chunkOut := range command.Generate(chunk) {
				response += chunkOut.Response
				out <- chunkOut
			}

			if shouldEnd := pipeline.next(out, response); shouldEnd {
				break
			}
		}
	}()

	return out
}

func (pipeline *Pipeline) next(out chan Chunk, response string) bool {
	errnie.Trace()

	commands := ExtractJSON(response)

	if len(commands) > 0 {
		var prompt string

		for _, command := range commands {
			switch command["command"] {
			case "transfer":
				prompt = command["arguments"].(map[string]any)["context"].(string)
			case "goto":
				teamName := command["arguments"].(map[string]any)["team"].(string)

				if prompt == "" {
					prompt = command["arguments"].(map[string]any)["prompt"].(string)
				}

				team := pipeline.teams[teamName]
				team.SetPrompt(prompt)
				chunk := pipeline.newChunk()
				chunk.Team = team

				for chunkOut := range team.Generate(chunk) {
					out <- chunkOut
				}

				prompt = ""
			case "return":
				pipeline.Response = command["response"].(string)
				return true
			default:
				os.Exit(1)
			}
		}
	}

	return false
}
