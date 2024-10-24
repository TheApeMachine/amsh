package mastercomputer

import (
	"fmt"
	"strings"

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

	page := instance.MustPage(
		args["url"].(string),
	).MustWindowFullscreen()

	// Wait for the page to load
	page.MustWaitStable()

	// Execute the provided JavaScript function
	result := page.MustEval(
		args["javascript"].(string),
	)

	out := strings.TrimSpace(strings.ReplaceAll(result.String(), "\n", " "))

	// Use some intelligent truncation to avoid overwhelming the message history
	if len(out) > 2000 {
		out = out[:2000]
	}

	return out, nil
}
