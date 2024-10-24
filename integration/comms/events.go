package comms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/theapemachine/amsh/data"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/twoface"
)

type Events struct {
	appToken string
	botToken string
	api      *slack.Client
	queue    *twoface.Queue
}

func NewEvents() *Events {
	botToken := os.Getenv("MARVIN_BOT_TOKEN")
	return &Events{
		appToken: os.Getenv("MARVIN_APP_TOKEN"),
		botToken: botToken,
		api:      slack.New(botToken),
		queue:    twoface.NewQueue(),
	}
}

func (srv *Events) Run(ctx fiber.Ctx) error {
	// signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	body, err := io.ReadAll(bytes.NewReader(ctx.Body()))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).SendString("failed to read request body")
	}

	// Convert the header to http.Header
	header := http.Header{}
	header.Add("X-Slack-Signature", string(ctx.Request().Header.Peek("X-Slack-Signature")))

	// sv, err := slack.NewSecretsVerifier(header, signingSecret)
	// if errnie.Error(err) != nil {
	// 	return ctx.Status(fiber.StatusBadRequest).SendString("failed to create secrets verifier")
	// }

	// if _, err := sv.Write(body); errnie.Error(err) != nil {
	// 	return ctx.Status(fiber.StatusInternalServerError).SendString("failed to write to secrets verifier")
	// }

	// if err := sv.Ensure(); errnie.Error(err) != nil {
	// 	return ctx.Status(fiber.StatusUnauthorized).SendString("failed to verify request signature")
	// }

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if errnie.Error(err) != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString("failed to parse event")
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var r *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &r); errnie.Error(err) != nil {
			return ctx.Status(fiber.StatusInternalServerError).SendString("failed to unmarshal challenge")
		}

		ctx.Type("text/plain")
		return ctx.Status(fiber.StatusOK).SendString(r.Challenge)
	case slackevents.CallbackEvent:
		srv.handleCallbackEvent(ctx, eventsAPIEvent)
	default:
		return ctx.Status(fiber.StatusBadRequest).SendString("unsupported event type")
	}

	return nil
}

func (srv *Events) handleCallbackEvent(ctx fiber.Ctx, eventsAPIEvent slackevents.EventsAPIEvent) {
	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		srv.handleAppMention(ev)
	case *slackevents.MessageEvent:
		srv.handleMessage(ev)
	case *slackevents.ReactionAddedEvent:
		srv.handleReactionAdded(ev)
	default:
		fmt.Printf("Unsupported event type: %s\n", innerEvent.Type)
	}
	ctx.Status(fiber.StatusOK)
	ctx.SendString("ok")
}

func (srv *Events) handleAppMention(ev *slackevents.AppMentionEvent) {
	_, _, err := srv.api.PostMessage(ev.Channel, slack.MsgOptionText("Hello! How can I assist you?", false))
	if err != nil {
		errnie.Error(err)
	}
}

func (srv *Events) handleMessage(ev *slackevents.MessageEvent) {
	// Add custom logic for handling regular messages
	if ev.User != "D07Q5CSP2MS" && ev.Text != "" {
		if user, err := srv.api.GetUserInfo(ev.User); errnie.Error(err) == nil {
			// Marshal the message
			buf, err := json.Marshal(ev)
			if errnie.Error(err) != nil {
				return
			}

			message := data.New(user.Profile.FirstName, "slack", "communicating", buf)
			message.Poke("chain", user.Profile.FirstName)
			srv.queue.PubCh <- *message
		}
	}
}

func (srv *Events) handleReactionAdded(ev *slackevents.ReactionAddedEvent) {
	// Add custom logic for handling reactions
	fmt.Printf("Reaction added: %s\n", ev.Reaction)
}
