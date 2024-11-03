package features

// import (
// 	"errors"
// 	"io/fs"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"sort"
// 	"strings"

// 	"github.com/theapemachine/amsh/errnie"
// 	"github.com/theapemachine/amsh/tui/core"
// )

// const (
// 	fileIconWidth = 3 // Width of "📁 " or "📄 "
// )

// type FileEntry struct {
// 	Name     string
// 	Path     string
// 	IsDir    bool
// 	Children []*FileEntry
// }

// // Browser implements core.BrowserInterface
// type Browser struct {
// 	root     *FileEntry
// 	current  *FileEntry
// 	workDir  string
// 	visible  bool
// 	selected int
// 	hasFocus bool
// }

// func NewBrowser() *Browser {
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		wd = "."
// 	}

// 	browser := &Browser{
// 		workDir:  wd,
// 		visible:  false,
// 		selected: 0,
// 	}

// 	// Initialize the browser
// 	if err := browser.Refresh(); err != nil {
// 		log.Printf("Error refreshing browser: %v", err)
// 	}
// 	return browser
// }

// // Toggle shows/hides the browser
// func (browser *Browser) Toggle() {
// 	browser.visible = !browser.visible
// 	browser.hasFocus = browser.visible
// 	if browser.visible {
// 		browser.Refresh()
// 	}
// }

// // IsVisible returns whether the browser is currently shown
// func (browser *Browser) IsVisible() bool {
// 	return browser.visible
// }

// // Refresh reloads the file tree
// func (browser *Browser) Refresh() error {
// 	root, err := browser.loadDir(browser.workDir)
// 	if err != nil {
// 		return errnie.Error(errors.New("failed to load root directory: " + browser.workDir))
// 	}

// 	browser.root = root
// 	browser.current = root // Make sure current is set to root initially

// 	// Add logging to verify the state
// 	errnie.Debug("Browser refreshed: root has %d children, current dir: %s",
// 		len(browser.root.Children), browser.current.Path)

// 	return nil
// }

// // loadDir recursively loads directory contents
// func (browser *Browser) loadDir(path string) (*FileEntry, error) {
// 	info, err := os.Stat(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	entry := &FileEntry{
// 		Name:  info.Name(),
// 		Path:  path,
// 		IsDir: info.IsDir(),
// 	}

// 	if !info.IsDir() {
// 		return entry, nil
// 	}

// 	files := errnie.SafeMust(func() ([]fs.DirEntry, error) { return os.ReadDir(path) })

// 	// Sort files: directories first, then files, both alphabetically
// 	sort.Slice(files, func(i, j int) bool {
// 		if files[i].IsDir() != files[j].IsDir() {
// 			return files[i].IsDir()
// 		}
// 		return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
// 	})

// 	for _, file := range files {
// 		// Skip hidden files
// 		if strings.HasPrefix(file.Name(), ".") {
// 			continue
// 		}

// 		childPath := filepath.Join(path, file.Name())
// 		child := errnie.SafeMust(func() (*FileEntry, error) { return browser.loadDir(childPath) })
// 		if child == nil {
// 			continue // Skip files we can't access
// 		}
// 		entry.Children = append(entry.Children, child)
// 	}

// 	return entry, nil
// }

// // MoveUp moves the selection up and updates the cursor
// func (browser *Browser) MoveUp(cursor *core.Cursor) {
// 	if !browser.HasFocus() {
// 		return
// 	}
// 	if browser.selected > 0 {
// 		browser.selected--
// 		cursor.MoveTo(browser.selected, fileIconWidth) // Position after the icon
// 	}
// }

// // MoveDown moves the selection down and updates the cursor
// func (browser *Browser) MoveDown(cursor *core.Cursor) {
// 	if !browser.HasFocus() {
// 		return
// 	}
// 	if browser.current != nil && browser.selected < len(browser.current.Children)-1 {
// 		browser.selected++
// 		cursor.MoveTo(browser.selected, fileIconWidth) // Position after the icon
// 	}
// }

// // Enter enters the selected directory or opens the selected file
// func (browser *Browser) Enter(cursor *core.Cursor) (string, error) {
// 	if !browser.visible || browser.current == nil || browser.selected >= len(browser.current.Children) {
// 		return "", nil
// 	}

// 	selected := browser.current.Children[browser.selected]
// 	if selected.IsDir {
// 		browser.current = selected
// 		browser.selected = 0
// 		cursor.MoveTo(0, fileIconWidth) // Position after the icon
// 		return "", nil
// 	}

// 	// Return the file path for loading
// 	return selected.Path, nil
// }

// // Back moves up one directory and updates the cursor
// func (browser *Browser) Back(cursor *core.Cursor) {
// 	if !browser.visible || browser.current == nil || browser.current == browser.root {
// 		return
// 	}

// 	if parent := findParent(browser.root, browser.current); parent != nil {
// 		browser.current = parent
// 		browser.selected = 0
// 		cursor.MoveTo(0, fileIconWidth) // Position after the icon
// 	}
// }

// // findParent finds the parent directory of the current directory
// func findParent(root, current *FileEntry) *FileEntry {
// 	if root == nil || current == nil {
// 		return nil
// 	}

// 	// Use a stack to perform a depth-first search
// 	stack := []*FileEntry{root}
// 	for len(stack) > 0 {
// 		node := stack[len(stack)-1]
// 		stack = stack[:len(stack)-1]

// 		for _, child := range node.Children {
// 			if child == current {
// 				return node
// 			}
// 			if child.IsDir {
// 				stack = append(stack, child)
// 			}
// 		}
// 	}

// 	return nil
// }

// // GetCurrentView returns the current directory listing for rendering
// func (browser *Browser) GetCurrentView() []string {
// 	if !browser.visible || browser.current == nil {
// 		return nil
// 	}

// 	var entries []string
// 	for _, entry := range browser.current.Children {
// 		prefix := "  "
// 		if entry.IsDir {
// 			prefix = "📁 "
// 		} else {
// 			prefix = "📄 "
// 		}
// 		entries = append(entries, prefix+entry.Name)
// 	}

// 	return entries
// }

// // GetSelected returns the index of the currently selected item
// func (browser *Browser) GetSelected() int {
// 	return browser.selected
// }

// // LoadFile reads the contents of a file
// func (browser *Browser) LoadFile(path string) ([]string, error) {
// 	content := errnie.SafeMust(func() ([]byte, error) { return os.ReadFile(path) })
// 	if content == nil {
// 		return nil, errnie.Error(errors.New("failed to read file: " + path))
// 	}

// 	// Split content into lines
// 	return strings.Split(string(content), "\n"), nil
// }

// func (browser *Browser) HasFocus() bool {
// 	return browser.visible && browser.hasFocus
// }

// func (browser *Browser) SetFocus(focus bool) {
// 	browser.hasFocus = focus
// }
