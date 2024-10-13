package comms

import "os"

type Search struct {
	appToken string
	botToken string
}

func NewSearch() *Search {
	return &Search{
		appToken: os.Getenv("APP_TOKEN"),
		botToken: os.Getenv("BOT_TOKEN"),
	}
}
