package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/slack-go/slack"
	"github.com/theapemachine/amsh/errnie"
)

type Slack struct {
	api *slack.Client
}

func NewSlack() *Slack {
	botToken := os.Getenv("MARVIN_BOT_TOKEN")

	if botToken == "" {
		errnie.Error(errors.New("BOT_TOKEN is not set"))
		return nil
	}

	return &Slack{
		api: slack.New(botToken),
	}
}

func (slack *Slack) GenerateSchema() string {
	schema := jsonschema.Reflect(&Slack{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

func (slack *Slack) Use(ctx context.Context, args map[string]any) string {
	return ""
}
