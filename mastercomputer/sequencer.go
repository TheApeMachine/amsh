package mastercomputer

import (
	"context"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/utils"
)

type Sequencer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	buffer   *data.Artifact
	executor *Executor
	name     string
	role     string
	scope    string
	worker   *Worker
	workers  map[string][]*Worker
	user     string
}

func NewSequencer(name string, role string, scope string, user string) *Sequencer {
	return &Sequencer{
		name:    name,
		role:    role,
		scope:   scope,
		workers: make(map[string][]*Worker),
		user:    user,
	}
}

func (sequencer *Sequencer) Initialize() {
	sequencer.ctx, sequencer.cancel = context.WithCancel(context.Background())
	sequencer.buffer = data.New(
		sequencer.name, sequencer.role, sequencer.scope, []byte(sequencer.user),
	)

	v := viper.GetViper()
	toolset := NewToolset()

	sequencer.executor = NewExecutor(sequencer)
	sequencer.worker = NewWorker(sequencer.ctx, sequencer.name, toolset.Assign(sequencer.role), sequencer.executor, sequencer.role)
	system := v.GetString("ai.prompt.system")
	system = strings.ReplaceAll(system, "{name}", sequencer.name)
	system = strings.ReplaceAll(system, "{role}", v.GetString("ai.prompt.sequencer.role"))
	sequencer.worker.system = system
	sequencer.worker.user = sequencer.user
	sequencer.worker.toolset = toolset.Assign(sequencer.role)
	sequencer.worker.Initialize()

	for _, role := range []string{"prompt", "reasoner", "researcher", "planner", "actor"} {
		for _, wrkr := range []string{utils.NewName(), utils.NewName(), utils.NewName()} {
			worker := NewWorker(sequencer.ctx, wrkr, toolset.Assign(role), sequencer.executor, role)
			system = v.GetString("ai.prompt.system")
			system = strings.ReplaceAll(system, "{name}", sequencer.name)
			system = strings.ReplaceAll(system, "{role}", v.GetString("ai.prompt."+role+".role"))
			worker.system = system
			worker.user = sequencer.user
			worker.Initialize()
			sequencer.workers[role] = append(sequencer.workers[role], worker)
		}
	}
}

func (sequencer *Sequencer) Start() {
	sequencer.worker.Start()
}

func (sequencer *Sequencer) Stop() {
	sequencer.cancel()
}
