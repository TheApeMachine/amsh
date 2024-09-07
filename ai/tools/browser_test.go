package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBrowser(t *testing.T) {
	Convey("Given a Browser instance", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body><h1>Test Page</h1><p>This is a test.</p></body></html>`))
		}))
		defer ts.Close()

		path, _ := launcher.LookPath()
		if path == "" {
			t.Log("Could not find Chrome executable")
			t.Log("Falling back to default launcher behavior")
		}

		tempDir := t.TempDir()
		launcher := launcher.New().
			Set("user-data-dir", tempDir).
			Delete("--headless")

		u := launcher.MustLaunch()
		browser := rod.New().ControlURL(u).MustConnect()
		
		defer func() {
			browser.MustClose()
			launcher.Cleanup()
			// Give some time for resources to be released
			time.Sleep(100 * time.Millisecond)
		}()

		b := &Browser{browser: browser}

		Convey("When using BrowseWeb", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			content, err := b.BrowseWeb(ctx, ts.URL)
			So(err, ShouldBeNil)
			So(content, ShouldContainSubstring, "Test Page")
			So(content, ShouldContainSubstring, "This is a test.")
		})

		// Add other test cases here...
	})
}
