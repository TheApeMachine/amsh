package mastercomputer

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/theapemachine/amsh/errnie"
)

type Browser struct {
	Function *openai.FunctionDefinition
}

func NewBrowser() *Browser {
	errnie.Trace()

	return &Browser{
		Function: &openai.FunctionDefinition{
			Name:        "browser",
			Description: "Use a fully functional web browser, and JavaScript for very flexible control.",
			Strict:      true,
			Parameters: jsonschema.Definition{
				Type:                 jsonschema.Object,
				AdditionalProperties: false,
				Properties: map[string]jsonschema.Definition{
					"url": {
						Type:        jsonschema.String,
						Description: "The url you want to visit.",
					},
					"javascript": {
						Type:        jsonschema.String,
						Description: "The javascript to run on the page (using the developer console). You MUST use a function and return value, for example: () => 'hello world'",
					},
				},
				Required: []string{"system", "user", "toolset"},
			},
		},
	}
}

func (browser *Browser) Run(ctx context.Context, parentID string, args map[string]any) (string, error) {
	l := launcher.New().Headless(false) // Change to true in production

	u, err := l.Launch()
	if errnie.Error(err) != nil {
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
