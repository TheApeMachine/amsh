package tools

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

/*
Browser provides a headless web browsing capability for AI Agents.
It implements the Tool interface, allowing it to be used within the AI system.
*/
type Browser struct {
	browser *rod.Browser
	page    *rod.Page
	content string
}

/*
NewBrowser initializes a new Browser instance with a connected Rod browser.
*/
func NewBrowser() *Browser {
	browser := rod.New().MustConnect()
	return &Browser{
		browser: browser,
	}
}

/*
Read satisfies the io.Reader interface, allowing content retrieval from the browser.
*/
func (b *Browser) Read(p []byte) (n int, err error) {
	if b.content == "" {
		return 0, io.EOF
	}
	n = copy(p, b.content)
	b.content = b.content[n:]
	return n, nil
}

/*
Write satisfies the io.Writer interface, enabling URL navigation in the browser.
*/
func (b *Browser) Write(p []byte) (n int, err error) {
	url := string(p)
	b.content = "" // Clear previous content

	page, err := b.browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return 0, fmt.Errorf("failed to create page: %w", err)
	}
	b.page = page

	if err := b.page.WaitLoad(); err != nil {
		return 0, fmt.Errorf("failed to load page: %w", err)
	}

	content, err := b.page.HTML()
	if err != nil {
		return 0, fmt.Errorf("failed to get page content: %w", err)
	}

	b.content = content

	return len(p), nil
}

/*
Close ensures proper cleanup of browser resources.
*/
func (b *Browser) Close() error {
	if b.page != nil {
		b.page.MustClose()
	}
	return b.browser.Close()
}

/*
Definition provides the OpenAI function definition for the Browser tool.
This allows the AI to understand how to use the browser capability.
*/
func (b *Browser) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "browse_web",
			Description: "Browse the web and return the content of a webpage",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"url": {
						Type:        jsonschema.String,
						Description: "The URL of the webpage to browse",
					},
				},
				Required: []string{"url"},
			},
		},
	}
}

/*
BrowseWeb performs the actual web browsing operation, returning the page content.
This method is the core functionality of the Browser tool.
*/
func (b *Browser) BrowseWeb(ctx context.Context, url string) (string, error) {
	page, err := b.browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.MustClose()

	if err := page.WaitLoad(); err != nil {
		return "", fmt.Errorf("failed to load page: %w", err)
	}

	// We ignore the HTML content as we're focusing on the text content
	text, err := page.Eval(`() => document.body.innerText`)
	if err != nil {
		return "", fmt.Errorf("failed to extract text content: %w", err)
	}

	return fmt.Sprintf("URL: %s\n\nContent:\n%s", url, strings.TrimSpace(text.Value.String())), nil
}
