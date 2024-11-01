package process

/*
CompositeProcess is a Process that is a composition of other Processes.
*/
type CompositeProcess struct {
	Layers []*Layer `json:"layers"`
}

/*
CompositeProcessMap is a map of CompositeProcesses by key.
*/
var CompositeProcessMap = map[string]*CompositeProcess{
	"task_analysis": {Layers: []*Layer{LayerMap["task_analysis"]}},
	"trengo":        {Layers: []*Layer{LayerMap["trengo"]}},
	"pull_request":  {Layers: []*Layer{LayerMap["pull_request"]}},
}

/*
Layer is a collection of Processes that are related to each other in a way
that allows their results to be combined in a meaningful way, and serve as
the input for the next layer, and can be run in parallel.
*/
type Layer struct {
	Processes []Process `json:"processes"`
}

/*
LayerMap finds a single layer of Processes by key.
*/
var LayerMap = map[string]*Layer{
	"task_analysis": {Processes: []Process{
		ProcessMap["breakdown"],
		ProcessMap["planning"],
	}},
	"trengo": {Processes: []Process{
		ProcessMap["trengo"],
	}},
	"pull_request": {Processes: []Process{
		ProcessMap["pull_request"],
	}},
	"abstract": {Processes: []Process{
		ProcessMap["surface"],
		ProcessMap["pattern"],
		ProcessMap["quantum"],
		ProcessMap["time"],
	}},
	"bridge": {Processes: []Process{
		ProcessMap["narrative"],
		ProcessMap["analogy"],
		ProcessMap["practical"],
		ProcessMap["context"],
	}},
	"ideate": {Processes: []Process{
		ProcessMap["moonshot"],
		ProcessMap["sensible"],
		ProcessMap["catalyst"],
		ProcessMap["guardian"],
	}},
}

/*
Process defines an interface that object can implement if the want to act
as a predefined process. Predefined processes are used to direct specific
behavior, useful is cases where we know what should be done based on an input.
*/
type Process interface {
	SystemPrompt(key string) string
}

/*
ProcessMap finds a single process by key, which is used to map incoming
WebHooks to the correct pre-defined process.
*/
var ProcessMap = map[string]Process{
	"breakdown": &Breakdown{},
	"planning":  &Planning{},
}
