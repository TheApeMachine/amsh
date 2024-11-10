package mastercomputer

import (
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/ai/boogie"
	"github.com/theapemachine/amsh/ai/process/quantum"
	"github.com/theapemachine/amsh/errnie"
)

type Prompt struct {
	systemPrompt string
	rolePrompt   string
	buffer       strings.Builder
}

func NewPrompt(role string) *Prompt {
	return &Prompt{
		systemPrompt: viper.GetViper().GetString("ai.setups.allied_mastercomputer.templates.system"),
		rolePrompt:   viper.GetViper().GetString("ai.setups.allied_mastercomputer.templates." + role),
	}
}

func (prompt *Prompt) Build(op boogie.Operation, state boogie.State) string {
	prompt.buffer.WriteString("Current operation: " + op.Name + "\n")
	prompt.buffer.WriteString("Available outcomes: " + strings.Join(op.Outcomes, ", ") + "\n\n")

	// Add context information
	prompt.buffer.WriteString("Current context:\n")
	contextJSON, _ := json.Marshal(state.Context)
	prompt.buffer.WriteString(string(contextJSON) + "\n\n")

	// Add operation-specific instructions
	switch {
	case strings.HasPrefix(op.Name, "quantum"):
		prompt.buffer.WriteString("This is a quantum process operation.\n")
		prompt.buffer.WriteString("Your response should be a valid quantum state update in JSON format.\n")

		if qstate, ok := state.Context["quantum_state"].(quantum.State); ok {
			currentState := errnie.SafeMust(func() ([]byte, error) {
				return json.Marshal(qstate)
			})

			prompt.buffer.WriteString("Current quantum state:\n" + string(currentState) + "\n")
		}

		// Add more process types here
	}

	return prompt.buffer.String()
}
