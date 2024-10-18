package comms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/theapemachine/amsh/errnie"
)

type Events struct {
	appToken string
	botToken string
	api      *slack.Client
}

func NewEvents() *Events {
	botToken := os.Getenv("BOT_TOKEN")
	return &Events{
		appToken: os.Getenv("APP_TOKEN"),
		botToken: botToken,
		api:      slack.New(botToken),
	}
}

func (srv *Events) Run(ctx context.Context) error {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	http.HandleFunc("/events-endpoint", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			http.Error(w, "Failed to create secrets verifier", http.StatusBadRequest)
			return
		}

		if _, err := sv.Write(body); err != nil {
			http.Error(w, "Failed to write to secrets verifier", http.StatusInternalServerError)
			return
		}

		if err := sv.Ensure(); err != nil {
			http.Error(w, "Failed to verify request signature", http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			http.Error(w, "Failed to parse event", http.StatusInternalServerError)
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			srv.handleURLVerification(w, body)
		case slackevents.CallbackEvent:
			srv.handleCallbackEvent(w, eventsAPIEvent)
		default:
			http.Error(w, "Unsupported event type", http.StatusBadRequest)
		}
	})

	errnie.Info("Slack Event Server listening on :8568")
	return http.ListenAndServe(":8568", nil)
}

func (srv *Events) handleURLVerification(w http.ResponseWriter, body []byte) {
	var r *slackevents.ChallengeResponse
	if err := json.Unmarshal(body, &r); err != nil {
		http.Error(w, "Failed to unmarshal challenge", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(r.Challenge))
}

func (srv *Events) handleCallbackEvent(w http.ResponseWriter, eventsAPIEvent slackevents.EventsAPIEvent) {
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
	w.WriteHeader(http.StatusOK)
}

func (srv *Events) handleAppMention(ev *slackevents.AppMentionEvent) {
	_, _, err := srv.api.PostMessage(ev.Channel, slack.MsgOptionText("Hello! How can I assist you?", false))
	if err != nil {
		errnie.Error(err)
	}
}

func (srv *Events) handleMessage(ev *slackevents.MessageEvent) {
	// Add custom logic for handling regular messages
	fmt.Printf("Received message: %s\n", ev.Text)
}

func (srv *Events) handleReactionAdded(ev *slackevents.ReactionAddedEvent) {
	// Add custom logic for handling reactions
	fmt.Printf("Reaction added: %s\n", ev.Reaction)
}
