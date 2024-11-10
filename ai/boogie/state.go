package boogie

// State represents the current execution state
type State struct {
	Context     map[string]interface{}
	CurrentStep string
	Outcome     string
	Error       error
}
