package mastercomputer

import (
	"context"
	"sync"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/utils"
)

type WorkerType string

const (
	WorkerTypeManager      WorkerType = "manager"
	WorkerTypeReasoner     WorkerType = "reasoner"
	WorkerTypeExecutor     WorkerType = "executor"
	WorkerTypeVerifier     WorkerType = "verifier"
	WorkerTypeCommunicator WorkerType = "communicator"
	WorkerTypeResearcher   WorkerType = "researcher"
	WorkerTypeWorker       WorkerType = "worker"
)

var builderInstance *Builder
var builderOnce sync.Once

type Builder struct {
	ctx     context.Context
	manager *Manager
}

func NewBuilder() *Builder {
	builderOnce.Do(func() {
		builderInstance = &Builder{ctx: context.Background(), manager: NewManager()}
	})
	return builderInstance
}

func (builder *Builder) NewWorker(t WorkerType, names ...string) *Worker {
	v := viper.GetViper()
	system := v.GetString("ai.prompt.system")
	role := v.GetString("ai.prompt." + string(t))
	guidelines := v.GetString("ai.prompt.guidelines")

	ID := utils.NewID()

	name := utils.NewName()
	if len(names) > 0 {
		name = names[0]
	}

	system = utils.ReplaceWith(system, [][]string{
		{"id", ID},
		{"name", name},
		{"role", role},
		{"guidelines", guidelines},
	})

	artifact := data.New(name, string(t), "system", nil)
	artifact.Poke("id", ID)
	artifact.Poke("system", system)
	artifact.Poke("workload", builder.GetWorkload(t))

	for key, value := range v.GetStringMapString("ai.config." + string(t)) {
		artifact.Poke(key, value)
	}

	return NewWorker(
		builder.ctx, artifact, builder.manager,
	).Initialize()
}

func (builder *Builder) GetRole(workload string) WorkerType {
	switch workload {
	case "reasoning":
		return WorkerTypeReasoner
	case "executing":
		return WorkerTypeExecutor
	case "managing":
		return WorkerTypeManager
	case "verifying":
		return WorkerTypeVerifier
	case "communicating":
		return WorkerTypeCommunicator
	case "researcher":
		return WorkerTypeResearcher
	default:
		return WorkerTypeWorker
	}
}

func (builder *Builder) GetWorkload(t WorkerType) string {
	switch t {
	case WorkerTypeManager:
		return "managing"
	case WorkerTypeReasoner:
		return "reasoning"
	case WorkerTypeExecutor:
		return "executing"
	case WorkerTypeVerifier:
		return "verifying"
	case WorkerTypeCommunicator:
		return "communicating"
	case WorkerTypeResearcher:
		return "researching"
	default:
		return "working"
	}
}

func (builder *Builder) Wait() {
	builder.manager.Wait()
}
