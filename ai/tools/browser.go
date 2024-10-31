package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/invopop/jsonschema"
	"github.com/spf13/cast"
	"github.com/theapemachine/amsh/errnie"
)

type BrowserArgs struct {
	URL        string `json:"url" jsonschema:"required,description=The URL to navigate to"`
	Selector   string `json:"selector" jsonschema:"description=CSS selector to find elements"`
	Timeout    int    `json:"timeout" jsonschema:"description=Timeout in seconds"`
	Screenshot bool   `json:"screenshot" jsonschema:"description=Whether to take a screenshot"`
}

type Browser struct {
	Operation  string            `json:"operation" jsonschema:"title=Operation,description=The operation to perform,enum=navigate,enum=click,enum=extract,enum=script,enum=wait,enum=form,enum=screenshot,enum=intercept,enum=cookies,enum=hijack,enum=response,enum=close"`
	Javascript string            `json:"javascript" jsonschema:"title=JavaScript,description=JavaScript code to execute in the developer console"`
	Hijack     string            `json:"hijack" jsonschema:"title=Hijack,description=Hijack a network request"`
	Response   string            `json:"response" jsonschema:"title=Response,description=Response to return for a network request"`
	Form       map[string]string `json:"form" jsonschema:"title=Form,description=Form data to fill in"`
	Intercept  []string          `json:"intercept" jsonschema:"title=Intercept,description=Network intercept patterns"`
	Cookies    string            `json:"cookies" jsonschema:"title=Cookies,description=Cookie operation,enum=get,enum=set,enum=delete"`
	instance   *rod.Browser
	page       *rod.Page
	history    []BrowseAction
	proxy      *url.URL
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

func (browser *Browser) Use(ctx context.Context, args map[string]any) string {
	result, err := browser.Run(args)
	if err != nil {
		errnie.Error(err)
	}
	return result
}

func (browser *Browser) GenerateSchema() string {
	schema := jsonschema.Reflect(&Browser{})
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errnie.Error(err)
	}
	return string(out)
}

// SetProxy configures a proxy for the browser
func (browser *Browser) SetProxy(proxyURL string) error {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}
	browser.proxy = proxy
	return nil
}

// StartSession initializes a new browsing session with stealth mode
func (browser *Browser) StartSession() error {
	l := launcher.New().
		Headless(false).
		Set("disable-web-security", "").
		Set("disable-setuid-sandbox", "").
		Set("no-sandbox", "")

	if browser.proxy != nil {
		l.Proxy(browser.proxy.String())
	}

	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	browser.instance = rod.New().
		ControlURL(url).
		MustConnect()

	// Create a new stealth page instead of regular page
	page, err := stealth.Page(browser.instance)
	if err != nil {
		return fmt.Errorf("failed to create stealth page: %w", err)
	}
	browser.page = page

	// Enable stealth mode and network interception
	browser.instance.MustIgnoreCertErrors(true)

	return nil
}

// Navigate goes to a URL and waits for the page to load
func (browser *Browser) Navigate(url string) error {
	// Instead of creating a new page, use the existing stealth page
	err := browser.page.Navigate(url)
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	err = browser.page.WaitLoad()
	if err != nil {
		return fmt.Errorf("failed to load page: %w", err)
	}

	browser.recordAction("navigate", url, "", err == nil)
	return nil
}

// Click finds and clicks an element using various selectors
func (browser *Browser) Click(selector string) error {
	el, err := browser.page.Element(selector)
	if err != nil {
		return fmt.Errorf("failed to find element: %w", err)
	}

	err = el.Click(proto.InputMouseButtonLeft, 1)
	browser.recordAction("click", selector, "", err == nil)
	return err
}

// Extract gets content from the page using a CSS selector
func (browser *Browser) Extract(selector string) (string, error) {
	el, err := browser.page.Element(selector)
	if err != nil {
		return "", fmt.Errorf("failed to find element: %w", err)
	}

	text, err := el.Text()
	browser.recordAction("extract", selector, text, err == nil)
	return text, err
}

// ExecuteScript runs custom JavaScript and returns the result
func (browser *Browser) ExecuteScript(script string) (interface{}, error) {
	result, err := browser.page.Eval(script)
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}

	browser.recordAction("script", script, fmt.Sprintf("%v", result), err == nil)
	return result.Value, nil
}

// WaitForElement waits for an element to appear
func (browser *Browser) WaitForElement(selector string, timeout time.Duration) error {
	// Remove unused context
	err := browser.page.Timeout(timeout).MustElement(selector).WaitVisible()
	browser.recordAction("wait", selector, "", err == nil)
	return err
}

// GetHistory returns the browsing session history
func (browser *Browser) GetHistory() []BrowseAction {
	return browser.history
}

// Run implements the enhanced interface with all new capabilities
func (browser *Browser) Run(args map[string]any) (string, error) {
	if proxyURL, ok := args["proxy"].(string); ok {
		if err := browser.SetProxy(proxyURL); err != nil {
			return "", err
		}
	}

	if err := browser.StartSession(); err != nil {
		return "", err
	}
	defer browser.instance.Close()

	if err := browser.Navigate(args["url"].(string)); err != nil {
		return "", err
	}

	// Handle form filling
	if formData, ok := args["form"].(map[string]string); ok {
		if err := browser.FillForm(formData); err != nil {
			return "", err
		}
	}

	// Handle screenshots
	if screenshot, ok := args["screenshot"].(map[string]string); ok {
		if err := browser.Screenshot(screenshot["selector"], screenshot["filepath"]); err != nil {
			return "", err
		}
	}

	// Handle network interception
	if patterns, ok := args["intercept"].([]string); ok {
		if err := browser.InterceptNetwork(patterns); err != nil {
			return "", err
		}
	}

	// Handle cookie operations
	if cookieOp, ok := args["cookies"].(string); ok {
		switch cookieOp {
		case "get":
			cookies, err := browser.ManageCookies()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%+v", cookies), nil
		case "set":
			if cookie, ok := args["cookie"].(Cookie); ok {
				if err := browser.SetCookie(cookie); err != nil {
					return "", err
				}
			}
		case "delete":
			if cookieData, ok := args["cookie_data"].(map[string]string); ok {
				if err := browser.DeleteCookies(cookieData["name"], cookieData["domain"]); err != nil {
					return "", err
				}
			}
		}
	}

	// Continue with existing functionality...
	if script, ok := args["javascript"].(string); ok {
		result, err := browser.ExecuteScript(script)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", result), nil
	}

	return browser.page.MustInfo().Title, nil
}

// FillForm fills form fields with provided data
func (browser *Browser) FillForm(fields map[string]string) error {
	for selector, value := range fields {
		el, err := browser.page.Element(selector)
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

		browser.recordAction("fill_form", map[string]string{
			"selector": selector,
			"value":    value,
		}, "", err == nil)
	}
	return nil
}

// Screenshot captures the current page or element
func (browser *Browser) Screenshot(selector string, filepath string) error {
	var img []byte
	var err error

	if selector == "" {
		// Capture full page
		img, err = browser.page.Screenshot(true, &proto.PageCaptureScreenshot{
			Format:      proto.PageCaptureScreenshotFormatPng,
			FromSurface: true,
		})

		if err != nil {
			return fmt.Errorf("failed to capture screenshot: %w", err)
		}
	} else {
		// Capture specific element
		el, err := browser.page.Element(selector)
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

	browser.recordAction("screenshot", map[string]string{
		"selector": selector,
		"filepath": filepath,
	}, "", err == nil)
	return nil
}

// InterceptNetwork starts intercepting network requests
func (browser *Browser) InterceptNetwork(patterns []string) error {
	router := browser.page.HijackRequests()
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

			browser.recordAction("network_request", req, "", true)
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		})
	}

	return nil
}

// ManageCookies provides cookie management capabilities
func (browser *Browser) ManageCookies() ([]Cookie, error) {
	cookies, err := browser.page.Cookies([]string{})
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

	browser.recordAction("get_cookies", nil, fmt.Sprintf("%d cookies", len(result)), err == nil)
	return result, nil
}

// SetCookie adds a new cookie
func (browser *Browser) SetCookie(cookie Cookie) error {
	// Fix SetCookies argument type
	err := browser.page.SetCookies([]*proto.NetworkCookieParam{
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

	browser.recordAction("set_cookie", cookie, "", err == nil)
	return err
}

// DeleteCookies removes cookies matching the given parameters
func (browser *Browser) DeleteCookies(name, domain string) error {
	// Fix non-variadic function call
	err := browser.page.SetCookies([]*proto.NetworkCookieParam{
		{
			Name:   name,
			Domain: domain,
		},
	})

	browser.recordAction("delete_cookies", map[string]string{
		"name":   name,
		"domain": domain,
	}, "", err == nil)
	return err
}

func (browser *Browser) recordAction(actionType string, data interface{}, result string, success bool) {
	browser.history = append(browser.history, BrowseAction{
		Type:    actionType,
		Data:    data,
		Result:  result,
		Time:    time.Now(),
		Success: success,
	})
}

func (browser *Browser) Close() error {
	if browser.instance != nil {
		return browser.instance.Close()
	}
	return nil
}

// Add the Execute method to implement the Tool interface
func (browser *Browser) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get URL from args
	url, ok := args["url"].(string)
	if !ok {
		return "", fmt.Errorf("url argument is required and must be a string")
	}

	// Navigate to URL
	if err := browser.Navigate(url); err != nil {
		return "", fmt.Errorf("failed to navigate: %w", err)
	}

	// Get selector if provided, default to "body"
	selector, err := getStringArg(args, "selector", "body")
	if err != nil {
		return "", fmt.Errorf("failed to get selector: %w", err)
	}

	// Wait for element if timeout provided
	if timeout, ok := args["timeout"].(float64); ok {
		if err := browser.WaitForElement(selector, time.Duration(timeout)*time.Second); err != nil {
			return "", fmt.Errorf("failed to find element: %w", err)
		}
	}

	// Extract content
	content, err := browser.Extract(selector)
	if err != nil {
		return "", fmt.Errorf("failed to extract content: %w", err)
	}

	return content, nil
}

func getStringArg(args map[string]interface{}, key string, defaultValue string) (string, error) {
	value, ok := args[key]
	if !ok {
		return defaultValue, nil
	}
	return cast.ToStringE(value)
}
