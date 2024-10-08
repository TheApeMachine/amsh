package mastercomputer

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/theapemachine/amsh/errnie"
)

type Toolset struct {
	tools map[string][]openai.Tool
}

func NewToolSet(ctx context.Context) *Toolset {
	errnie.Trace()

	return &Toolset{
		tools: map[string][]openai.Tool{
			"control": {
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewWorker(ctx).Function,
				},
				openai.Tool{
					Type:     openai.ToolTypeFunction,
					Function: NewCommand().Function,
				},
			},
		},
	}
}

func (toolset *Toolset) Tools(key string) []openai.Tool {
	errnie.Trace()

	return toolset.tools[key]
}
