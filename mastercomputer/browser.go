package mastercomputer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/theapemachine/amsh/berrt"
)

type Browser struct {
	url        string
	javascript string
}

func NewBrowser(parameters map[string]any) *Browser {
	url, urlOk := parameters["url"].(string)
	javascript, jsOk := parameters["javascript"].(string)

	if !urlOk || !jsOk {
		berrt.Error("browser", errors.New("invalid parameters for browser"))
		return nil
	}

	return &Browser{
		url:        url,
		javascript: javascript,
	}
}

func (browser *Browser) Start() string {
	if browser == nil {
		berrt.Error("browser", errors.New("browser instance is nil"))
		return ""
	}

	l := launcher.New().Headless(false) // Change to true in production

	u, err := l.Launch()
	if err != nil {
		berrt.Error("browser", err)
		return ""
	}

	defer l.Cleanup()

	instance := rod.New().ControlURL(u).MustConnect()
	defer func() {
		if err := instance.Close(); err != nil {
			fmt.Printf("Error closing browser: %v\n", err)
		}
	}()

	page := instance.MustPage(
		browser.url,
	).MustWindowFullscreen()

	// Wait for the page to load
	page.MustWaitStable()

	// Execute the provided JavaScript function
	result := page.MustEval(
		browser.javascript,
	)

	out := strings.TrimSpace(strings.ReplaceAll(result.String(), "\n", " "))

	// Use some intelligent truncation to avoid overwhelming the message history
	if len(out) > 2000 {
		out = out[:2000]
	}

	berrt.Info("browser", out)

	return out
}
