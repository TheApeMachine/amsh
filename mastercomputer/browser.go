package mastercomputer

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

	page := instance.MustPage(
		args["url"].(string),
	).MustWindowFullscreen()

	// Wait for the page to load
	page.MustWaitStable()

	// Execute the provided JavaScript function
	result := page.MustEval(
		args["javascript"].(string),
	)

	fmt.Println("[BROWSER RESULT]")
	fmt.Println(result.String())
	fmt.Println("[/BROWSER RESULT]")

	return result.String(), nil
}
