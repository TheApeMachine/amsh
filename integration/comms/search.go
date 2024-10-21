package comms

import (
	"context"
	"os"

	"github.com/slack-go/slack"
)

type Search struct {
	appToken string
	botToken string
	api      *slack.Client
}

func NewSearch() *Search {
	botToken := os.Getenv("MARVIN_BOT_TOKEN")
	userToken := os.Getenv("MARVIN_USER_TOKEN")
	return &Search{
		appToken: os.Getenv("MARVIN_APP_TOKEN"),
		botToken: botToken,
		api:      slack.New(userToken),
	}
}

func (s *Search) SearchMessages(ctx context.Context, query string) ([]slack.SearchMessage, error) {
	params := slack.NewSearchParameters()
	results, err := s.api.SearchMessagesContext(ctx, query, params)
	if err != nil {
		return nil, err
	}
	return results.Matches, nil
}

func (s *Search) SearchFiles(ctx context.Context, query string) ([]slack.File, error) {
	params := slack.NewSearchParameters()
	results, err := s.api.SearchFilesContext(ctx, query, params)
	if err != nil {
		return nil, err
	}
	return results.Matches, nil
}
