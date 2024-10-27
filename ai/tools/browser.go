package tools

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/spf13/cast"
	"github.com/theapemachine/amsh/ai/types"
)

type Browser struct {
	Instance *rod.Browser // Changed from instance to Instance to make it public
	page     *rod.Page
	history  []BrowseAction
	proxy    *url.URL
}

type BrowseAction struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Result  string      `json:"result"`
	Time    time.Time   `json:"time"`
	Success bool        `json:"success"`
}

type NetworkRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Cookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HTTPOnly bool      `json:"http_only"`
}

func NewBrowser() *Browser {
	return &Browser{
		history: make([]BrowseAction, 0),
	}
}

// SetProxy configures a proxy for the browser
func (b *Browser) SetProxy(proxyURL string) error {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}
	b.proxy = proxy
	return nil
}

// StartSession initializes a new browsing session with stealth mode
func (b *Browser) StartSession() error {
	l := launcher.New().
		Headless(false).
		Set("disable-web-security", "").
		Set("disable-setuid-sandbox", "").
		Set("no-sandbox", "")

	if b.proxy != nil {
		l.Proxy(b.proxy.String())
	}

	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	b.Instance = rod.New().
		ControlURL(url).
		MustConnect()

	// Create a new stealth page instead of regular page
	page, err := stealth.Page(b.Instance)
	if err != nil {
		return fmt.Errorf("failed to create stealth page: %w", err)
	}
	b.page = page

	// Enable stealth mode and network interception
	b.Instance.MustIgnoreCertErrors(true)

	return nil
}

// Navigate goes to a URL and waits for the page to load
func (b *Browser) Navigate(url string) error {
	// Instead of creating a new page, use the existing stealth page
	err := b.page.Navigate(url)
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	err = b.page.WaitLoad()
	if err != nil {
		return fmt.Errorf("failed to load page: %w", err)
	}

	b.recordAction("navigate", url, "", err == nil)
	return nil
}

// Click finds and clicks an element using various selectors
func (b *Browser) Click(selector string) error {
	el, err := b.page.Element(selector)
	if err != nil {
		return fmt.Errorf("failed to find element: %w", err)
	}

	err = el.Click(proto.InputMouseButtonLeft, 1)
	b.recordAction("click", selector, "", err == nil)
	return err
}

// Extract gets content from the page using a CSS selector
func (b *Browser) Extract(selector string) (string, error) {
	el, err := b.page.Element(selector)
	if err != nil {
		return "", fmt.Errorf("failed to find element: %w", err)
	}

	text, err := el.Text()
	b.recordAction("extract", selector, text, err == nil)
	return text, err
}

// ExecuteScript runs custom JavaScript and returns the result
func (b *Browser) ExecuteScript(script string) (interface{}, error) {
	result, err := b.page.Eval(script)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	b.recordAction("script", script, fmt.Sprintf("%v", result), err == nil)
	return result.Value, nil
}

// WaitForElement waits for an element to appear
func (b *Browser) WaitForElement(selector string, timeout time.Duration) error {
	// Remove unused context
	err := b.page.Timeout(timeout).MustElement(selector).WaitVisible()
	b.recordAction("wait", selector, "", err == nil)
	return err
}

// GetHistory returns the browsing session history
func (b *Browser) GetHistory() []BrowseAction {
	return b.history
}

// Run implements the enhanced interface with all new capabilities
func (b *Browser) Run(args map[string]any) (string, error) {
	if proxyURL, ok := args["proxy"].(string); ok {
		if err := b.SetProxy(proxyURL); err != nil {
			return "", err
		}
	}

	if err := b.StartSession(); err != nil {
		return "", err
	}
	defer b.Instance.Close()

	if err := b.Navigate(args["url"].(string)); err != nil {
		return "", err
	}

	// Handle form filling
	if formData, ok := args["form"].(map[string]string); ok {
		if err := b.FillForm(formData); err != nil {
			return "", err
		}
	}

	// Handle screenshots
	if screenshot, ok := args["screenshot"].(map[string]string); ok {
		if err := b.Screenshot(screenshot["selector"], screenshot["filepath"]); err != nil {
			return "", err
		}
	}

	// Handle network interception
	if patterns, ok := args["intercept"].([]string); ok {
		if err := b.InterceptNetwork(patterns); err != nil {
			return "", err
		}
	}

	// Handle cookie operations
	if cookieOp, ok := args["cookies"].(string); ok {
		switch cookieOp {
		case "get":
			cookies, err := b.ManageCookies()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%+v", cookies), nil
		case "set":
			if cookie, ok := args["cookie"].(Cookie); ok {
				if err := b.SetCookie(cookie); err != nil {
					return "", err
				}
			}
		case "delete":
			if cookieData, ok := args["cookie_data"].(map[string]string); ok {
				if err := b.DeleteCookies(cookieData["name"], cookieData["domain"]); err != nil {
					return "", err
				}
			}
		}
	}

	// Continue with existing functionality...
	if script, ok := args["javascript"].(string); ok {
		result, err := b.ExecuteScript(script)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", result), nil
	}

	return b.page.MustInfo().Title, nil
}

// FillForm fills form fields with provided data
func (b *Browser) FillForm(fields map[string]string) error {
	for selector, value := range fields {
		el, err := b.page.Element(selector)
		if err != nil {
			return fmt.Errorf("failed to find form field %s: %w", selector, err)
		}

		// Clear existing value
		el.MustEval(`el => el.value = ''`)

		// Input new value
		err = el.Input(value)
		if err != nil {
			return fmt.Errorf("failed to fill field %s: %w", selector, err)
		}

		b.recordAction("fill_form", map[string]string{
			"selector": selector,
			"value":    value,
		}, "", err == nil)
	}
	return nil
}

// Screenshot captures the current page or element
func (b *Browser) Screenshot(selector string, filepath string) error {
	var img []byte
	var err error

	if selector == "" {
		// Capture full page
		img, err = b.page.Screenshot(true, &proto.PageCaptureScreenshot{
			Format:      proto.PageCaptureScreenshotFormatPng,
			FromSurface: true,
		})

		if err != nil {
			return fmt.Errorf("failed to capture screenshot: %w", err)
		}
	} else {
		// Capture specific element
		el, err := b.page.Element(selector)
		if err != nil {
			return fmt.Errorf("failed to find element for screenshot: %w", err)
		}
		img, err = el.Screenshot(proto.PageCaptureScreenshotFormatPng, 1)
		if err != nil {
			return fmt.Errorf("failed to capture screenshot: %w", err)
		}
	}

	err = os.WriteFile(filepath, img, 0644)
	if err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	b.recordAction("screenshot", map[string]string{
		"selector": selector,
		"filepath": filepath,
	}, "", err == nil)
	return nil
}

// InterceptNetwork starts intercepting network requests
func (b *Browser) InterceptNetwork(patterns []string) error {
	router := b.page.HijackRequests()
	defer router.Stop()

	for _, pattern := range patterns {
		router.MustAdd(pattern, func(ctx *rod.Hijack) {
			// Fix headers type conversion
			headers := make(map[string]string)
			for k, v := range ctx.Request.Headers() {
				headers[k] = v.String() // Convert gson.JSON to string
			}

			// Fix body type conversion
			body := ctx.Request.Body() // Call the function to get the body string

			req := NetworkRequest{
				URL:     ctx.Request.URL().String(),
				Method:  ctx.Request.Method(),
				Headers: headers,
				Body:    body,
			}

			b.recordAction("network_request", req, "", true)
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
	}

	return nil
}

// ManageCookies provides cookie management capabilities
func (b *Browser) ManageCookies() ([]Cookie, error) {
	cookies, err := b.page.Cookies([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cookies: %w", err)
	}

	var result []Cookie
	for _, c := range cookies {
		cookie := Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  time.Unix(int64(c.Expires), 0),
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
		}
		result = append(result, cookie)
	}

	b.recordAction("get_cookies", nil, fmt.Sprintf("%d cookies", len(result)), err == nil)
	return result, nil
}

// SetCookie adds a new cookie
func (b *Browser) SetCookie(cookie Cookie) error {
	// Fix SetCookies argument type
	err := b.page.SetCookies([]*proto.NetworkCookieParam{
		{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  proto.TimeSinceEpoch(cookie.Expires.Unix()),
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
		},
	})

	b.recordAction("set_cookie", cookie, "", err == nil)
	return err
}

// DeleteCookies removes cookies matching the given parameters
func (b *Browser) DeleteCookies(name, domain string) error {
	// Fix non-variadic function call
	err := b.page.SetCookies([]*proto.NetworkCookieParam{
		{
			Name:   name,
			Domain: domain,
		},
	})

	b.recordAction("delete_cookies", map[string]string{
		"name":   name,
		"domain": domain,
	}, "", err == nil)
	return err
}

func (b *Browser) recordAction(actionType string, data interface{}, result string, success bool) {
	b.history = append(b.history, BrowseAction{
		Type:    actionType,
		Data:    data,
		Result:  result,
		Time:    time.Now(),
		Success: success,
	})
}

func (b *Browser) Close() error {
	if b.Instance != nil {
		return b.Instance.Close()
	}
	return nil
}

// Add the Execute method to implement the Tool interface
func (b *Browser) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get URL from args
	url, ok := args["url"].(string)
	if !ok {
		return "", fmt.Errorf("url argument is required and must be a string")
	}

	// Navigate to URL
	if err := b.Navigate(url); err != nil {
		return "", fmt.Errorf("failed to navigate: %w", err)
	}

	// Get selector if provided, default to "body"
	selector, err := getStringArg(args, "selector", "body")
	if err != nil {
		return "", fmt.Errorf("failed to get selector: %w", err)
	}

	// Wait for element if timeout provided
	if timeout, ok := args["timeout"].(float64); ok {
		if err := b.WaitForElement(selector, time.Duration(timeout)*time.Second); err != nil {
			return "", fmt.Errorf("failed to find element: %w", err)
		}
	}

	// Extract content
	content, err := b.Extract(selector)
	if err != nil {
		return "", fmt.Errorf("failed to extract content: %w", err)
	}

	return content, nil
}

// GetSchema implements the Tool interface
func (b *Browser) GetSchema() types.ToolSchema {
	return types.ToolSchema{
		Name:        "browser",
		Description: "Web browser automation tool for navigating and extracting content",
		Parameters: map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to navigate to",
				"required":    true,
			},
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector to extract content from",
				"default":     "body",
			},
			"timeout": map[string]interface{}{
				"type":        "number",
				"description": "Timeout in seconds for waiting for elements",
				"default":     5,
			},
		},
	}
}

func getStringArg(args map[string]interface{}, key string, defaultValue string) (string, error) {
	value, ok := args[key]
	if !ok {
		return defaultValue, nil
	}
	return cast.ToStringE(value)
}

// TestStealth checks if our stealth configuration is working
func (b *Browser) TestStealth() error {
	// Navigate to the bot detection test site
	if err := b.Navigate("https://bot.sannysoft.com"); err != nil {
		return fmt.Errorf("failed to navigate to test site: %w", err)
	}

	// Wait for results to load
	if err := b.WaitForElement("#broken-image-dimensions.passed", 10*time.Second); err != nil {
		return fmt.Errorf("failed to load test results: %w", err)
	}

	// Extract and log results
	results, err := b.Extract("table")
	if err != nil {
		return fmt.Errorf("failed to extract results: %w", err)
	}

	log.Printf("Stealth Test Results:\n%s", results)

	// Optionally save a screenshot
	if err := b.Screenshot("", "stealth_test.png"); err != nil {
		log.Printf("Failed to save screenshot: %v", err)
	}

	return nil
}
