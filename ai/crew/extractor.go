package crew

import (
	"context"
	"strings"

	"github.com/theapemachine/amsh/errnie"

	"github.com/spf13/viper"
)

/*
Extractor interface defines the behavior of the Extractor agent.
*/
type Extractor interface {
	UpdateProfile(agent *Agent) error
}

type extractor struct {
	agent *Agent
}

/*
NewExtractor initializes a new Extractor agent.
*/
func NewExtractor(ctx context.Context) Extractor {
	return &extractor{
		agent: NewAgent(ctx, NewConn(), "extractor", Colors[3]),
	}
}

func (e *extractor) UpdateProfile(agent *Agent) error {
	system := viper.GetString("ai.crew.extractor.system")
	user := viper.GetString("ai.crew.extractor.user")

	user = strings.ReplaceAll(user, "<{profile}>", agent.Profile.String())
	user = strings.ReplaceAll(user, "<{history}>", agent.history)

	content, err := e.agent.ChatCompletion(system, user)
	if err != nil {
		errnie.Error(err.Error())
		return err
	}

	agent.Profile.Unmarshal(ExtractJSON(content))
	return nil
}
