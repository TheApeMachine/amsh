package comms

import (
	"context"
	"os"

	"github.com/slack-go/slack"
	"github.com/theapemachine/amsh/errnie"
)

type Message struct {
	appToken string
	botToken string
	channel  string
	pretext  string
	text     string
}

func NewMessage(channel, pretext, text string) *Message {
	return &Message{
		appToken: os.Getenv("APP_TOKEN"),
		botToken: os.Getenv("BOT_TOKEN"),
		channel:  channel,
		pretext:  pretext,
		text:     text,
	}
}

func (msg *Message) Post(ctx context.Context) (err error) {
	api := slack.New(msg.botToken)

	if _, _, err = api.PostMessage(
		msg.channel,
		slack.MsgOptionText(msg.text, false),
	); err != nil {
		return errnie.Error(err)
	}

	return
}

func (msg *Message) ConversationHistory(ctx context.Context) (messages []slack.Message, err error) {
	api := slack.New(msg.botToken)

	params := slack.GetConversationHistoryParameters{
		ChannelID: msg.channel,
	}

	var response *slack.GetConversationHistoryResponse
	if response, err = api.GetConversationHistoryContext(ctx, &params); err != nil {
		return nil, errnie.Error(err)
	}

	messages = append(messages, response.Messages...)
	return
}
