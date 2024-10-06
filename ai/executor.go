package ai

type Executor interface {
	AddAgents(agents ...*Agent)
	Execute(system string, user string)
	Pause()
}
