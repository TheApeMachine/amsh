package types

import (
	"os"
	"path/filepath"
	"sort"
)

// FileEntry represents a file or directory in the browser
type FileEntry struct {
	Name     string
	IsDir    bool
	FullPath string
}

// Browser represents a file browser state
type Browser struct {
	currentPath string
	entries     []FileEntry
	selected    int
	scrollY     int
}

// NewBrowser creates a new file browser
func NewBrowser() *Browser {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}
	b := &Browser{
		currentPath: pwd,
		selected:    0,
		scrollY:     0,
	}
	b.Refresh()
	return b
}

// Refresh reloads the current directory listing
func (b *Browser) Refresh() error {
	entries, err := os.ReadDir(b.currentPath)
	if err != nil {
		return err
	}

	// Convert to our FileEntry type
	b.entries = make([]FileEntry, 0, len(entries))

	// Add parent directory if not at root
	if b.currentPath != "/" {
		b.entries = append(b.entries, FileEntry{
			Name:     "..",
			IsDir:    true,
			FullPath: filepath.Join(b.currentPath, ".."),
		})
	}

	// Add all other entries
	for _, entry := range entries {
		// Skip hidden files
		if entry.Name()[0] == '.' && entry.Name() != ".." {
			continue
		}

		b.entries = append(b.entries, FileEntry{
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			FullPath: filepath.Join(b.currentPath, entry.Name()),
		})
	}

	// Sort entries: directories first, then files, both alphabetically
	sort.Slice(b.entries, func(i, j int) bool {
		if b.entries[i].IsDir != b.entries[j].IsDir {
			return b.entries[i].IsDir
		}
		return b.entries[i].Name < b.entries[j].Name
	})

	return nil
}

// MoveUp moves the selection up
func (b *Browser) MoveUp() {
	if b.selected > 0 {
		b.selected--
	}
}

// MoveDown moves the selection down
func (b *Browser) MoveDown() {
	if b.selected < len(b.entries)-1 {
		b.selected++
	}
}

// Enter enters the selected directory or returns the selected file
func (b *Browser) Enter() (string, bool, error) {
	if b.selected >= len(b.entries) {
		return "", false, nil
	}

	entry := b.entries[b.selected]
	if entry.IsDir {
		b.currentPath = entry.FullPath
		b.selected = 0
		b.scrollY = 0
		err := b.Refresh()
		return "", true, err
	}

	return entry.FullPath, false, nil
}

// GetEntries returns the current entries
func (b *Browser) GetEntries() []FileEntry {
	return b.entries
}

// GetSelected returns the selected index
func (b *Browser) GetSelected() int {
	return b.selected
}

// GetScroll returns the scroll position
func (b *Browser) GetScroll() int {
	return b.scrollY
}

// SetScroll sets the scroll position
func (b *Browser) SetScroll(scroll int) {
	b.scrollY = scroll
}

// GetCurrentPath returns the current directory path
func (b *Browser) GetCurrentPath() string {
	return b.currentPath
}
