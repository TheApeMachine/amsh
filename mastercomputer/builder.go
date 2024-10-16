package mastercomputer

import (
	"context"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/utils"
)

type WorkerType string

const (
	WorkerTypeManager  WorkerType = "manager"
	WorkerTypeReasoner WorkerType = "reasoner"
	WorkerTypeExecutor WorkerType = "executor"
)

type Builder struct {
	ctx     context.Context
	manager *WorkerManager
}

func NewBuilder(ctx context.Context, manager *WorkerManager) *Builder {
	return &Builder{ctx: ctx, manager: manager}
}

func (builder *Builder) Worker(t WorkerType) *Worker {
	v := viper.GetViper()
	system := v.GetString("ai.prompt.system")
	role := v.GetString("ai.prompt." + string(t))
	guidelines := v.GetString("ai.prompt.guidelines")

	system = utils.ReplaceWith(system, [][]string{
		{"role", role},
		{"guidelines", guidelines},
	})

	artifact := data.New(utils.NewID(), string(t), "system", nil)
	artifact.Poke("system", system)
	artifact.Poke("workload", builder.getWorkload(t))

	return NewWorker(
		builder.ctx, artifact, builder.manager,
	).Initialize()
}

func (builder *Builder) getWorkload(t WorkerType) string {
	switch t {
	case WorkerTypeManager:
		return "managing"
	case WorkerTypeReasoner:
		return "reasoning"
	case WorkerTypeExecutor:
		return "executing"
	default:
		return ""
	}
}
