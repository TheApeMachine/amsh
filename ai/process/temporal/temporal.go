package temporal

import (
	"time"

	"github.com/theapemachine/amsh/utils"
)

/*
Process represents the evolution of thoughts over time.
*/
type Process struct {
	Timeline       []TimePoint     `json:"timeline" jsonschema:"title=Timeline,description=Sequence of thought states; required"`
	CausalChains   []CausalChain   `json:"causal_chains" jsonschema:"title=CausalChains,description=Cause-effect relationships over time; required"`
	EvolutionRules []EvolutionRule `json:"evolution_rules" jsonschema:"title=EvolutionRules,description=Patterns of state change; required"`
}

func NewProcess() *Process {
	return &Process{}
}

type CausalChain struct {
	ID       string     `json:"id" jsonschema:"required,title=ID,description=Unique identifier for causal chain"`
	EventIDs []string   `json:"event_ids" jsonschema:"required,title=EventIDs,description=IDs of events in chain"`
	Strength float64    `json:"strength" jsonschema:"required,title=Strength,description=Causal relationship strength"`
	Evidence []Evidence `json:"evidence" jsonschema:"required,title=Evidence,description=Supporting evidence"`
}

type Evidence struct {
	Type        string  `json:"type" jsonschema:"required,title=Type,description=Type of evidence"`
	Description string  `json:"description" jsonschema:"required,title=Description,description=Evidence description"`
	Confidence  float64 `json:"confidence" jsonschema:"required,title=Confidence,description=Confidence level"`
	Source      string  `json:"source" jsonschema:"required,title=Source,description=Evidence source"`
}

type TimePoint struct {
	Time   time.Time              `json:"time" jsonschema:"required,title=Time,description=Point in time"`
	State  map[string]interface{} `json:"state" jsonschema:"required,title=State,description=System state"`
	Delta  map[string]float64     `json:"delta" jsonschema:"required,title=Delta,description=State changes"`
	Events []Event                `json:"events" jsonschema:"required,title=Events,description=Events at this time"`
}

type Event struct {
	ID        string                 `json:"id" jsonschema:"required,title=ID,description=Unique identifier for event"`
	Type      string                 `json:"type" jsonschema:"required,title=Type,description=Type of event"`
	Data      map[string]interface{} `json:"data" jsonschema:"description=Event data"`
	Timestamp time.Time              `json:"timestamp" jsonschema:"description=Event time"`
}

type EvolutionRule struct {
	ID          string    `json:"id" jsonschema:"required,description=Unique identifier for the evolution rule"`
	Condition   Predicate `json:"condition" jsonschema:"required,description=Condition for the evolution rule"`
	Action      Transform `json:"action" jsonschema:"required,description=Action to be taken"`
	Priority    int       `json:"priority" jsonschema:"required,description=Priority of the evolution rule"`
	Reliability float64   `json:"reliability" jsonschema:"required,description=Reliability of the evolution rule"`
}

type Predicate struct {
	Type      string                 `json:"type" jsonschema:"required,description=Type of the predicate"`
	Params    map[string]interface{} `json:"params" jsonschema:"required,description=Parameters for the predicate"`
	Threshold float64                `json:"threshold" jsonschema:"required,description=Threshold for the predicate"`
}

type Transform struct {
	Type      string                 `json:"type" jsonschema:"required,description=Type of the transformation"`
	Params    map[string]interface{} `json:"params" jsonschema:"required,description=Parameters for the transformation"`
	Magnitude float64                `json:"magnitude" jsonschema:"required,description=Magnitude of the transformation"`
}

func (ta *Process) GenerateSchema() string {
	return utils.GenerateSchema[Process]()
}
