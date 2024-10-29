package system

/*
Architecture determines the way the system components work together that
ultimately defines its behavior.
*/
type Architecture struct {
	Name           string          `json:"name"`
	ProcessManager *ProcessManager `json:"process_manager"`
}

/*
NewArchitecture creates a new instance of the specified architecture.
*/
func NewArchitecture(key string) *Architecture {
	return &Architecture{
		Name:           key,
		ProcessManager: NewProcessManager(key, key),
	}
}
