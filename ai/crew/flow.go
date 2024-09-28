package crew

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
Flow interface defines the behavior of the Flow agent.
*/
type Flow interface {
	Decide(action string) (*FlowDecision, error)
}

type flow struct {
	agent *Agent
}

/*
NewFlow initializes a new Flow agent.
*/
func NewFlow(ctx context.Context) Flow {
	return &flow{
		agent: NewAgent(ctx, NewConn(), "flow", Colors[2]),
	}
}

func (f *flow) Decide(action string) (*FlowDecision, error) {
	system := viper.GetString("ai.crew.flow.system")
	user := viper.GetString("ai.crew.flow.user")

	user = strings.ReplaceAll(user, "<{action}>", action)

	resp, err := f.agent.ChatCompletion(system, user)
	if err != nil {
		errnie.Error(err.Error())
		return nil, err
	}

	out := &FlowDecision{}
	buf := ExtractJSON(resp)

	if err = json.Unmarshal([]byte(buf), out); err != nil {
		errnie.Error(err.Error())
		return nil, err
	}

	return out, nil
}
