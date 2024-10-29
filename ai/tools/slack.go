package tools

import (
	"errors"
	"os"

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

func (slack *Slack) Use(args map[string]any) string {
	return ""
}
