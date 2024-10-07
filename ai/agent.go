package ai

import (
	"context"

	"github.com/theapemachine/amsh/ai/memory"
	"github.com/theapemachine/amsh/ai/tools"
)

type Agent struct {
	ctx     context.Context
	conn    *Conn
	ID      string
	Type    string
	memory  *memory.Memory
	Toolset []tools.Tool
	Active  bool
}

func NewAgent(
	ctx context.Context, conn *Conn, ID, Type string, toolset []tools.Tool,
) *Agent {
	return &Agent{
		ctx:     ctx,
		conn:    conn,
		ID:      ID,
		Type:    Type,
		memory:  memory.NewMemory(ID),
		Toolset: toolset,
		Active:  false,
	}
}

func (agent *Agent) Generate(ctx context.Context, prompt *Prompt) chan string {
	if !agent.Active {
		return nil
	}

	return agent.conn.Next(ctx, prompt)
}

func (agent *Agent) Start() {
	agent.Active = true
}

func (agent *Agent) Stop() {
	agent.Active = false
}
