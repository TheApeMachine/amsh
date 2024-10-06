package ai

import (
	"context"

	"github.com/theapemachine/amsh/errnie"
)

type Chunk struct {
	SessionID  string  `json:"session_id"`
	SequenceID string  `json:"sequence_id"`
	Iteration  int     `json:"iteration"`
	Prompt     *Prompt `json:"prompt"`
	Agent      string  `json:"agent"`
	AgentType  string  `json:"agent_type"`
	Response   string  `json:"response"`
}

type Pipeline struct {
	ctx              context.Context
	conn             *Conn
	agents           map[string]*Agent
	resourceManager  *ResourceManager
	history          string
	approachHistory  []string
	currentIteration int
	maxIterations    int
	agentResponses   map[string]string
}

func NewPipeline(ctx context.Context, conn *Conn) *Pipeline {
	return &Pipeline{
		ctx:              ctx,
		conn:             conn,
		agents:           make(map[string]*Agent),
		resourceManager:  NewResourceManager(),
		approachHistory:  make([]string, 0),
		currentIteration: 0,
	}
}

func (pipeline *Pipeline) Initialize() *Pipeline {
	errnie.Trace()
	return pipeline
}

func (pipeline *Pipeline) Generate(prompt string, iterations int) <-chan Chunk {
	return out
}
