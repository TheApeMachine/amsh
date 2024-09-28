package crew

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
Writer interface defines the behavior of the Writer agent.
*/
type Writer interface {
	Write(scene string, directions *Direction) (*Script, error)
}

type writer struct {
	agent *Agent
}

/*
NewWriter initializes a new Writer agent.
*/
func NewWriter(ctx context.Context) Writer {
	return &writer{
		agent: NewAgent(ctx, NewConn(), "writer", Colors[1]),
	}
}

func (w *writer) Write(scene string, directions *Direction) (*Script, error) {
	system := viper.GetString("ai.crew.writer.system")
	user := viper.GetString("ai.crew.writer.user")

	user = strings.ReplaceAll(user, "<{scene}>", scene)
	user = strings.ReplaceAll(user, "<{directions}>", directions.Description)

	resp, err := w.agent.ChatCompletion(system, user)
	if err != nil {
		errnie.Error(err.Error())
		return nil, err
	}

	out := &Script{}
	buf := ExtractJSON(resp)

	if err = json.Unmarshal([]byte(buf), out); err != nil {
		errnie.Error(err.Error())
		return nil, err
	}

	return out, nil
}
