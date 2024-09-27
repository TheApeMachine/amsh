package ai

import (
	"context"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

type Pipeline struct {
	ctx      context.Context
	conn     *Conn
	prompt   *Prompt
	steps    []string
	step     int
	out      chan string
	loglevel string
}

func NewPipeline(ctx context.Context, conn *Conn, steps ...string) *Pipeline {
	errnie.Debug("NewPipeline %v", steps)
	return &Pipeline{
		ctx:      ctx,
		conn:     conn,
		prompt:   NewPrompt("pipeline"),
		steps:    steps,
		step:     0,
		out:      make(chan string),
		loglevel: viper.GetViper().GetString("loglevel"),
	}
}

func (pipeline *Pipeline) AddTask(goal string) {
	errnie.Debug("AddTask %s", goal)
	pipeline.prompt = NewPrompt("pipeline")
}

func (pipeline *Pipeline) Generate() <-chan string {
	errnie.Debug("pipeline.Generate %v", pipeline.steps)

	go func() {
		defer close(pipeline.out)

		pipeline.initialize()

		for _, context := range pipeline.prompt.contexts {
			context.Agent.prompt.AddSystem(
				pipeline.prompt.systems[pipeline.step],
			).AddContext(context)

			errnie.Debug("pipeline.Generate %v", context.Agent.prompt)

			for chunk := range context.Agent.Generate(
				pipeline.ctx,
				pipeline.step,
			) {
				errnie.Debug("pipeline.Generate %v", chunk)
				pipeline.out <- chunk
			}
		}
	}()

	return pipeline.out
}

func (pipeline *Pipeline) initialize() {
	errnie.Debug("pipeline.initialize %v", pipeline.steps)
	ps := pipeline.steps[pipeline.step]
	pipeline.prompt.AddSystem(viper.GetViper().GetString("ai.prompt.system." + ps))

	for i := 0; i < 10; i++ {
		pipeline.prompt.AddContext(
			NewContext(
				NewAgent(
					pipeline.ctx,
					pipeline.conn,
					NewPrompt(ps),
				),
			),
		)
	}
}
