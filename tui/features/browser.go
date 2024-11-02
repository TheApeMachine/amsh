package features

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileEntry struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileEntry
}

// Browser implements core.BrowserInterface
type Browser struct {
	root     *FileEntry
	current  *FileEntry
	workDir  string
	visible  bool
	selected int
}

func NewBrowser() *Browser {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	browser := &Browser{
		workDir:  wd,
		visible:  false,
		selected: 0,
	}

	browser.Refresh()
	return browser
}

// Toggle shows/hides the browser
func (browser *Browser) Toggle() {
	browser.visible = !browser.visible
	if browser.visible {
		browser.Refresh()
	}
}

// IsVisible returns whether the browser is currently shown
func (browser *Browser) IsVisible() bool {
	return browser.visible
}

// Refresh reloads the file tree
func (browser *Browser) Refresh() error {
	root, err := browser.loadDir(browser.workDir)
	if err != nil {
		return err
	}

	browser.root = root
	browser.current = root
	return nil
}

// loadDir recursively loads directory contents
func (browser *Browser) loadDir(path string) (*FileEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	entry := &FileEntry{
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return entry, nil
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Sort files: directories first, then files, both alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() != files[j].IsDir() {
			return files[i].IsDir()
		}
		return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
	})

	for _, file := range files {
		// Skip hidden files
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		childPath := filepath.Join(path, file.Name())
		child, err := browser.loadDir(childPath)
		if err != nil {
			continue // Skip files we can't access
		}
		entry.Children = append(entry.Children, child)
	}

	return entry, nil
}

// MoveUp moves the selection up
func (browser *Browser) MoveUp() {
	log.Printf("Browser MoveUp called, current selected: %d", browser.selected)
	if browser.selected > 0 {
		browser.selected--
	}
	log.Printf("New selected: %d", browser.selected)
}

// MoveDown moves the selection down
func (browser *Browser) MoveDown() {
	log.Printf("Browser MoveDown called, current selected: %d", browser.selected)
	if browser.current != nil && browser.selected < len(browser.current.Children)-1 {
		browser.selected++
	}
	log.Printf("New selected: %d, total items: %d", browser.selected, len(browser.current.Children))
}

// Enter enters the selected directory or opens the selected file
func (browser *Browser) Enter() (string, error) {
	log.Printf("Browser Enter called")
	if !browser.visible || browser.current == nil || browser.selected >= len(browser.current.Children) {
		log.Printf("Enter conditions not met: visible=%v, current=%v, selected=%d",
			browser.visible, browser.current != nil, browser.selected)
		return "", nil
	}

	selected := browser.current.Children[browser.selected]
	if selected.IsDir {
		browser.current = selected
		browser.selected = 0
		return "", nil
	}

	// Return the file path for loading
	return selected.Path, nil
}

// Back moves up one directory
func (browser *Browser) Back() {
	if !browser.visible || browser.current == nil || browser.current == browser.root {
		return
	}

	// Find parent directory
	var findParent func(*FileEntry, *FileEntry) *FileEntry
	findParent = func(root, target *FileEntry) *FileEntry {
		for _, child := range root.Children {
			if child == target {
				return root
			}
			if child.IsDir {
				if found := findParent(child, target); found != nil {
					return found
				}
			}
		}
		return nil
	}

	if parent := findParent(browser.root, browser.current); parent != nil {
		browser.current = parent
		browser.selected = 0
	}
}

// GetCurrentView returns the current directory listing for rendering
func (browser *Browser) GetCurrentView() []string {
	if !browser.visible || browser.current == nil {
		return nil
	}

	var entries []string
	for _, entry := range browser.current.Children {
		prefix := "  "
		if entry.IsDir {
			prefix = "📁 "
		} else {
			prefix = "📄 "
		}
		entries = append(entries, prefix+entry.Name)
	}
	return entries
}

// GetSelected returns the index of the currently selected item
func (browser *Browser) GetSelected() int {
	return browser.selected
}

// LoadFile reads the contents of a file
func (browser *Browser) LoadFile(path string) ([]string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Split content into lines
	return strings.Split(string(content), "\n"), nil
}
