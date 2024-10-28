package system

import (
	"fmt"

	"github.com/spf13/viper"
)

/*
Architecture determines the way the system components work together that
ultimately defines its behavior.
*/
type Architecture struct {
	Name           string          `json:"name"`
	SystemPrompt   string          `json:"system_prompt"`
	ProcessManager *ProcessManager `json:"process_manager"`
}

/*
NewArchitecture creates a new instance of the specified architecture.
*/
func NewArchitecture(key string) *Architecture {
	return &Architecture{
		Name:           key,
		SystemPrompt:   viper.GetString(fmt.Sprintf("ai.setups.%s.system", key)),
		ProcessManager: NewProcessManager(key),
	}
}
