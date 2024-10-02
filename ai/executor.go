package ai

import (
	"context"

	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tweaker"
)

var reset = "\033[0m"

type Executor struct {
	ctx     context.Context
	conn    *Conn
	setup   map[string]interface{}
	agents  []*Agent
	history string
}

func NewExecutor(ctx context.Context, conn *Conn) *Executor {
	errnie.Trace()

	return &Executor{
		ctx:     ctx,
		conn:    conn,
		setup:   tweaker.Setups(),
		agents:  make([]*Agent, 0),
		history: "",
	}
}

func (executor *Executor) Generate() <-chan string {
	errnie.Trace()

	out := make(chan string)

	go func() {
		defer close(out)

		for _, agent := range executor.agents {
			out <- agent.Color
			executor.history += "\n\n---\n\nAGENT: " + agent.ID + "\n\n"
			out <- agent.system + "\n\n" + agent.user + "\n\n"
			for chunk := range agent.Generate(executor.ctx, executor.history) {
				executor.history += chunk
				out <- chunk
			}
			out <- reset

			biasDetectionResult := executor.detectBias(executor.history)
			out <- "\n\nBias Detection: " + biasDetectionResult + "\n\n"
		}
	}()

	return out
}

func (executor *Executor) detectBias(text string) string {
	// Implement bias detection logic here
	// This is a placeholder implementation
	return "No significant biases detected."
}

func (executor *Executor) AddAgent(agent *Agent) {
	errnie.Trace()
	executor.agents = append(executor.agents, agent)
}
