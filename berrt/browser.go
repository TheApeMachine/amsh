package berrt

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type Browser struct {
}

func NewBrowser() *Browser {
	return &Browser{}
}

func (browser *Browser) Run(args map[string]any) (string, error) {
	l := launcher.New().Headless(false) // Change to true in production

	u, err := l.Launch()
	if err != nil {
		return "", err
	}

	defer l.Cleanup()

	instance := rod.New().ControlURL(u).MustConnect()
	defer func() {
		if err := instance.Close(); err != nil {
			fmt.Printf("Error closing browser: %v\n", err)
		}
	}()

	return instance.MustPage(
		args["url"].(string),
	).MustWindowFullscreen().MustEval(
		args["javascript"].(string),
	).Get("output").Str(), nil
}
