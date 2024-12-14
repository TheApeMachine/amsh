package comms

import (
	"context"
	"os"

	"github.com/slack-go/slack"
	"github.com/theapemachine/errnie"
)

type Message struct {
	appToken string
	botToken string
	channel  string
	text     string
	api      *slack.Client
}

func NewMessage(channel, text string) *Message {
	botToken := os.Getenv("BOT_TOKEN")
	return &Message{
		appToken: os.Getenv("APP_TOKEN"),
		botToken: botToken,
		channel:  channel,
		text:     text,
		api:      slack.New(botToken),
	}
}

func (msg *Message) Post(ctx context.Context) error {
	_, _, err := msg.api.PostMessageContext(
		ctx,
		msg.channel,
		slack.MsgOptionText(msg.text, false),
	)
	if err != nil {
		return errnie.Error(err)
	}
	return nil
}

func (msg *Message) PostWithAttachments(ctx context.Context, attachments []slack.Attachment) error {
	_, _, err := msg.api.PostMessageContext(
		ctx,
		msg.channel,
		slack.MsgOptionText(msg.text, false),
		slack.MsgOptionAttachments(attachments...),
	)
	if err != nil {
		return errnie.Error(err)
	}
	return nil
}

func (msg *Message) ConversationHistory(ctx context.Context) ([]slack.Message, error) {
	params := slack.GetConversationHistoryParameters{
		ChannelID: msg.channel,
		Limit:     100,
	}

	response, err := msg.api.GetConversationHistoryContext(ctx, &params)
	if err != nil {
		return nil, errnie.Error(err)
	}

	return response.Messages, nil
}

func (msg *Message) UpdateMessage(ctx context.Context, timestamp, newText string) error {
	_, _, _, err := msg.api.UpdateMessageContext(
		ctx,
		msg.channel,
		timestamp,
		slack.MsgOptionText(newText, false),
	)
	if err != nil {
		return errnie.Error(err)
	}
	return nil
}

func (msg *Message) DeleteMessage(ctx context.Context, timestamp string) error {
	_, _, err := msg.api.DeleteMessageContext(ctx, msg.channel, timestamp)
	if err != nil {
		return errnie.Error(err)
	}
	return nil
}
