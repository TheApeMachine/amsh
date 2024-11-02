package core

// BrowserInterface defines the methods needed by Screen to render a browser
type BrowserInterface interface {
	IsVisible() bool
	GetCurrentView() []string
	GetSelected() int
}
