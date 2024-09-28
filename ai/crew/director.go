package crew

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
	"github.com/theapemachine/amsh/errnie"
)

/*
Director interface defines the behavior of the Director agent.
*/
type Director interface {
	Direct(highlights string) (*Direction, error)
}

type director struct {
	agent *Agent
}

/*
NewDirector initializes a new Director agent.
*/
func NewDirector(ctx context.Context) Director {
	return &director{
		agent: NewAgent(ctx, NewConn(), "director", Colors[0]),
	}
}

func (d *director) Direct(highlights string) (*Direction, error) {
	system := viper.GetString("ai.crew.director.system")
	user := viper.GetString("ai.crew.director.user")

	user = strings.ReplaceAll(user, "<{highlights}>", highlights)

	resp, err := d.agent.ChatCompletion(system, user)
	if err != nil {
		errnie.Error(err.Error())
		return nil, err
	}

	out := &Direction{}
	buf := ExtractJSON(resp)

	if err = json.Unmarshal([]byte(buf), out); err != nil {
		errnie.Error(err.Error())
		return nil, err
	}

	return out, nil
}
