package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

type Pipeline struct {
	ctx    context.Context
	conn   *Conn
	roles  []string
	agents []*Agent
	prompt *Prompt
}

func NewPipeline(ctx context.Context, conn *Conn, roles ...string) *Pipeline {
	errnie.Debug("NewPipeline %v", roles)
	return &Pipeline{ctx: ctx, conn: conn, roles: roles}
}

func (pipeline *Pipeline) AddTask(goal string) {
	errnie.Debug("AddTask %s", goal)
	pipeline.prompt = NewPrompt("pipeline")
	pipeline.prompt.context = "## Goal\n\n" + goal + "\n\n## Context\n\n"
}

func (pipeline *Pipeline) Generate() <-chan string {
	errnie.Debug("Generate")

	out := make(chan string)

	// Create agents for each role
	for _, role := range pipeline.roles {
		name := namegenerator.NewNameGenerator(time.Now().UnixNano()).Generate()
		prompt := NewPrompt(role)

		pipeline.agents = append(
			pipeline.agents,
			NewAgent(
				pipeline.ctx,
				pipeline.conn,
				name,
				role,
				prompt,
			))
	}

	go func() {
		defer close(out)

		for _, agent := range pipeline.agents {
			steps := viper.GetStringSlice(fmt.Sprintf("ai.prompt.steps.%s", agent.role))

			out <- pipeline.replacements(pipeline.makeSystem(agent), agent)

			for _, step := range steps {
				agent.prompt.context = fmt.Sprintf("%s\n\n### Task\n\n> %s\n\n---\n\n### Response\n\n", pipeline.prompt.context, step)
				out <- agent.prompt.context

				for chunk := range agent.Generate(pipeline.ctx, pipeline.prompt.context) {
					pipeline.prompt.context += chunk
					out <- chunk
				}

				out <- "\n\n**Signed** " + agent.name + "\n\n---\n\n"
			}
		}
	}()

	return out
}

func (pipeline *Pipeline) makeSystem(agent *Agent) string {
	return strings.Join([]string{
		agent.prompt.system,
	}, "\n\n")
}

func (pipeline *Pipeline) replacements(prompt string, agent *Agent) string {
	prompt = strings.ReplaceAll(prompt, "<{profile}>", agent.prompt.role)
	prompt = strings.ReplaceAll(prompt, "<{name}>", "`"+agent.name+"`")
	prompt = strings.ReplaceAll(prompt, "<{modules}>", agent.prompt.modules)

	return prompt
}
