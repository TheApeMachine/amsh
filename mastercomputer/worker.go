package mastercomputer

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Worker struct {
	pctx     context.Context
	ctx      context.Context
	cancel   context.CancelFunc
	Function *openai.FunctionDefinition
}

func NewWorker(pctx context.Context) *Worker {
	errnie.Trace()

	return &Worker{
		pctx: pctx,
		Function: &openai.FunctionDefinition{
			Name:        "worker",
			Description: "Use to create a worker agent, pass in the system prompt and user prompt",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Description:          "Use to create a worker agent, pass in the system prompt and user prompt",
				Properties: map[string]jsonschema.Definition{
					"system": {
						Type:        jsonschema.String,
						Description: "The system prompt",
					},
					"user": {
						Type:        jsonschema.String,
						Description: "The user prompt",
					},
				},
				Required: []string{"system", "user"},
			},
		},
	}
}

func (worker *Worker) Initialize() {
	errnie.Trace()

	worker.ctx, worker.cancel = context.WithCancel(worker.pctx)
}

func (worker *Worker) Run() chan string {
	errnie.Trace()

	out := make(chan string)

	go func() {
		for {
			select {
			case <-worker.pctx.Done():
				worker.cancel()
				return
			default:
			}
		}
	}()

	return out
}
