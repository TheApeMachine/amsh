package ai

import (
	"context"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tweaker"
)

/*
Pipeline orchestrates the execution of scenes and agents.
*/
type Pipeline struct {
	ctx      context.Context
	conn     *Conn
	executor *Executor
}

/*
NewPipeline initializes the pipeline with agents and scenes.
*/
func NewPipeline(
	ctx context.Context,
	conn *Conn,
) *Pipeline {
	errnie.Trace()

	return &Pipeline{
		ctx:      ctx,
		conn:     NewConn(),
		executor: NewExecutor(ctx, conn),
	}
}

/*
Initialize the pipeline.
*/
func (pipeline *Pipeline) Initialize() *Pipeline {
	errnie.Trace()

	// Initialize the executor.
	pipeline.executor.Initialize()

	// Add the agents to the executor.
	for idx, agent := range tweaker.Agents() {
		pipeline.executor.AddAgent(NewAgent(
			pipeline.ctx,
			pipeline.conn,
			namegenerator.NewNameGenerator(
				time.Now().UTC().UnixNano(),
			).Generate(),
			agent["type"].(string),
			agent["scope"].(string),
			agent["responsibilities"].(string),
			Colors[idx%len(Colors)],
		))
	}

	return pipeline
}

/*
Generate the pipeline.
*/
func (pipeline *Pipeline) Generate() <-chan string {
	errnie.Trace()

	out := make(chan string)

	go func() {
		defer close(out)

		for chunk := range pipeline.executor.Generate() {
			out <- chunk
		}
	}()

	return out
}

/*
Save the pipeline.
*/
func (pipeline *Pipeline) Save() {
	errnie.Trace()
}
