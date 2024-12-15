package types

import (
	"os"
	"sort"
	"strings"
	"unicode"
)

// Buffer represents a text buffer with content and cursor position
type Buffer struct {
	lines           [][]rune
	cursorX         int
	cursorY         int
	scrollY         int
	filename        string
	selectionStart  Position            // Selection start position
	selectionEnd    Position            // Selection end position
	selectionActive bool                // Whether selection is active
	clipboard       string              // Store clipboard content
	jumpPoints      map[string]Position // For quick navigation
	jumpMode        bool
	jumpInput       string // Collect characters for jump label
}

// Position represents a position in the buffer
type Position struct {
	X int
	Y int
}

// NewBuffer creates a new empty buffer
func NewBuffer() *Buffer {
	return &Buffer{
		lines:           [][]rune{{}},
		selectionActive: false,
		clipboard:       "",
		jumpPoints:      make(map[string]Position),
		jumpMode:        false,
		jumpInput:       "",
	}
}

func (b *Buffer) MoveCursorLeft() {
	if b.cursorX > 0 {
		b.cursorX--
	}
}

func (b *Buffer) MoveCursorRight() {
	if b.cursorY < len(b.lines) && b.cursorX < len(b.lines[b.cursorY]) {
		b.cursorX++
	}
}

func (b *Buffer) MoveCursorUp() {
	if b.cursorY > 0 {
		b.cursorY--
		if b.cursorX > len(b.lines[b.cursorY]) {
			b.cursorX = len(b.lines[b.cursorY])
		}
	}
}

func (b *Buffer) MoveCursorDown() {
	if b.cursorY < len(b.lines)-1 {
		b.cursorY++
		if b.cursorX > len(b.lines[b.cursorY]) {
			b.cursorX = len(b.lines[b.cursorY])
		}
	}
}

func (b *Buffer) InsertRune(r rune) {
	line := b.lines[b.cursorY]
	if b.cursorX == len(line) {
		b.lines[b.cursorY] = append(line, r)
	} else {
		b.lines[b.cursorY] = append(line[:b.cursorX], append([]rune{r}, line[b.cursorX:]...)...)
	}
	b.cursorX++
}

func (b *Buffer) InsertNewline() {
	line := b.lines[b.cursorY]
	newLine := make([]rune, len(line[b.cursorX:]))
	copy(newLine, line[b.cursorX:])
	b.lines[b.cursorY] = line[:b.cursorX]
	b.lines = append(b.lines[:b.cursorY+1], append([][]rune{newLine}, b.lines[b.cursorY+1:]...)...)
	b.cursorY++
	b.cursorX = 0
}

func (b *Buffer) Backspace() {
	if b.cursorX > 0 {
		line := b.lines[b.cursorY]
		b.lines[b.cursorY] = append(line[:b.cursorX-1], line[b.cursorX:]...)
		b.cursorX--
	} else if b.cursorY > 0 {
		// Join with previous line
		prevLine := b.lines[b.cursorY-1]
		b.cursorX = len(prevLine)
		b.lines[b.cursorY-1] = append(prevLine, b.lines[b.cursorY]...)
		b.lines = append(b.lines[:b.cursorY], b.lines[b.cursorY+1:]...)
		b.cursorY--
	}
}

// LoadFile loads a file into the buffer
func (b *Buffer) LoadFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	b.filename = filename
	b.lines = [][]rune{{}}
	b.cursorX = 0
	b.cursorY = 0
	b.scrollY = 0

	currentLine := []rune{}
	for _, ch := range data {
		if ch == '\n' {
			b.lines = append(b.lines, currentLine)
			currentLine = []rune{}
		} else {
			currentLine = append(currentLine, rune(ch))
		}
	}
	if len(currentLine) > 0 {
		b.lines = append(b.lines, currentLine)
	}

	return nil
}

// GetLine returns the line at the given index
func (b *Buffer) GetLine(idx int) []rune {
	if idx >= 0 && idx < len(b.lines) {
		return b.lines[idx]
	}
	return nil
}

// GetFilename returns the buffer's filename
func (b *Buffer) GetFilename() string {
	return b.filename
}

// GetCursor returns the current cursor position
func (b *Buffer) GetCursor() (x, y int) {
	return b.cursorX, b.cursorY
}

// GetScroll returns the current scroll position
func (b *Buffer) GetScroll() int {
	return b.scrollY
}

// LineCount returns the total number of lines in the buffer
func (b *Buffer) LineCount() int {
	return len(b.lines)
}

// SetFilename sets the buffer's filename
func (b *Buffer) SetFilename(filename string) {
	b.filename = filename
}

// StartSelection starts a new selection at the current cursor position
func (b *Buffer) StartSelection() {
	b.selectionStart = Position{X: b.cursorX, Y: b.cursorY}
	b.selectionEnd = b.selectionStart
	b.selectionActive = true
}

// UpdateSelection updates the selection end point to the current cursor position
func (b *Buffer) UpdateSelection() {
	if b.selectionActive {
		b.selectionEnd = Position{X: b.cursorX, Y: b.cursorY}
	}
}

// ClearSelection clears the current selection
func (b *Buffer) ClearSelection() {
	b.selectionActive = false
}

// GetSelection returns the current selection bounds and whether selection is active
func (b *Buffer) GetSelection() (start Position, end Position, active bool) {
	if !b.selectionActive {
		return Position{}, Position{}, false
	}

	// Normalize selection bounds (start should be before end)
	start = b.selectionStart
	end = b.selectionEnd
	if start.Y > end.Y || (start.Y == end.Y && start.X > end.X) {
		start, end = end, start
	}
	return start, end, true
}

// IsPositionSelected returns whether a given position is within the selection
func (b *Buffer) IsPositionSelected(x, y int) bool {
	if !b.selectionActive {
		return false
	}

	start, end, _ := b.GetSelection()
	pos := Position{X: x, Y: y}

	// Check if position is within selection bounds
	if pos.Y < start.Y || pos.Y > end.Y {
		return false
	}
	if pos.Y == start.Y && pos.X < start.X {
		return false
	}
	if pos.Y == end.Y && pos.X > end.X {
		return false
	}
	return true
}

// GetSelectedText returns the currently selected text
func (b *Buffer) GetSelectedText() string {
	if !b.selectionActive {
		return ""
	}

	start, end, _ := b.GetSelection()
	var text strings.Builder

	// Single line selection
	if start.Y == end.Y {
		line := b.lines[start.Y]
		for i := start.X; i <= end.X && i < len(line); i++ {
			text.WriteRune(line[i])
		}
		return text.String()
	}

	// Multi-line selection
	// First line
	if start.Y < len(b.lines) {
		line := b.lines[start.Y]
		for i := start.X; i < len(line); i++ {
			text.WriteRune(line[i])
		}
		text.WriteRune('\n')
	}

	// Middle lines
	for y := start.Y + 1; y < end.Y && y < len(b.lines); y++ {
		text.WriteString(string(b.lines[y]))
		text.WriteRune('\n')
	}

	// Last line
	if end.Y < len(b.lines) {
		line := b.lines[end.Y]
		for i := 0; i <= end.X && i < len(line); i++ {
			text.WriteRune(line[i])
		}
	}

	return text.String()
}

// Yank copies the selected text to the clipboard
func (b *Buffer) Yank() string {
	text := b.GetSelectedText()
	b.clipboard = text
	return text
}

// Paste inserts the clipboard content at the current cursor position
func (b *Buffer) Paste() {
	if b.clipboard == "" {
		return
	}

	// Split clipboard content into lines
	lines := strings.Split(b.clipboard, "\n")

	// Handle first line - insert at cursor position
	firstLine := []rune(lines[0])
	currentLine := b.lines[b.cursorY]
	b.lines[b.cursorY] = append(currentLine[:b.cursorX], append(firstLine, currentLine[b.cursorX:]...)...)
	b.cursorX += len(firstLine)

	// Handle remaining lines
	if len(lines) > 1 {
		// Save the remainder of the current line
		remainder := b.lines[b.cursorY][b.cursorX:]

		// Insert middle lines
		for i := 1; i < len(lines)-1; i++ {
			newLine := []rune(lines[i])
			b.lines = append(b.lines[:b.cursorY+i], append([][]rune{newLine}, b.lines[b.cursorY+i:]...)...)
		}

		// Insert last line with remainder
		lastLine := []rune(lines[len(lines)-1])
		lastLine = append(lastLine, remainder...)
		b.lines = append(b.lines[:b.cursorY+len(lines)-1], append([][]rune{lastLine}, b.lines[b.cursorY+len(lines)-1:]...)...)

		// Update cursor position
		b.cursorY += len(lines) - 1
		b.cursorX = len(lines[len(lines)-1])
	}
}

// PasteLines inserts the clipboard content as new lines starting at the current line
func (b *Buffer) PasteLines() {
	if b.clipboard == "" {
		return
	}

	// Split clipboard content into lines
	lines := strings.Split(b.clipboard, "\n")

	// Insert each line as a new line
	for i, line := range lines {
		newLine := []rune(line)
		b.lines = append(b.lines[:b.cursorY+i], append([][]rune{newLine}, b.lines[b.cursorY+i:]...)...)
	}

	// Update cursor position
	b.cursorY += len(lines)
	b.cursorX = 0
}

// GetClipboard returns the current clipboard content
func (b *Buffer) GetClipboard() string {
	return b.clipboard
}

// SetClipboard sets the clipboard content
func (b *Buffer) SetClipboard(content string) {
	b.clipboard = content
}

// StartJumpMode calculates jump points and returns the labels
func (b *Buffer) StartJumpMode() map[string]Position {
	b.jumpMode = true
	b.jumpInput = ""
	b.jumpPoints = make(map[string]Position)

	// Use a map to deduplicate positions
	uniquePoints := make(map[Position]bool)

	// Find all potential jump points
	for y, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		// First char of line is always a jump point
		uniquePoints[Position{X: 0, Y: y}] = true

		inWord := false
		for x := 0; x < len(line); x++ {
			ch := line[x]

			// Skip consecutive spaces
			if ch == ' ' {
				inWord = false
				continue
			}

			// Start of word after space
			if !inWord {
				uniquePoints[Position{X: x, Y: y}] = true
				inWord = true
				continue
			}

			// CamelCase transition
			if x > 0 && unicode.IsLower(rune(line[x-1])) && unicode.IsUpper(rune(ch)) {
				uniquePoints[Position{X: x, Y: y}] = true
				continue
			}

			// After symbols (but not spaces)
			if x > 0 && !unicode.IsLetter(rune(line[x-1])) && !unicode.IsNumber(rune(line[x-1])) &&
				line[x-1] != ' ' && (unicode.IsLetter(rune(ch)) || unicode.IsNumber(rune(ch))) {
				uniquePoints[Position{X: x, Y: y}] = true
			}
		}
	}

	// Convert unique points to a slice for sorting
	var points []Position
	for pos := range uniquePoints {
		points = append(points, pos)
	}

	// Sort points by position (top to bottom, left to right)
	sort.Slice(points, func(i, j int) bool {
		if points[i].Y != points[j].Y {
			return points[i].Y < points[j].Y
		}
		return points[i].X < points[j].X
	})

	// Generate labels efficiently
	labels := generateSmartLabels(len(points))

	// Assign labels to points
	for i, pos := range points {
		if i < len(labels) {
			b.jumpPoints[labels[i]] = pos
		}
	}

	return b.jumpPoints
}

// generateSmartLabels generates a sequence of labels optimized for quick typing
func generateSmartLabels(count int) []string {
	var labels []string

	// First, use home row keys (asdfjkl;)
	homeRow := []rune{'a', 's', 'd', 'f', 'j', 'k', 'l'}
	for _, r := range homeRow {
		labels = append(labels, string(r))
		if len(labels) >= count {
			return labels
		}
	}

	// Then use other common keys
	commonKeys := []rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p',
		'g', 'h', 'z', 'x', 'c', 'v', 'b', 'n', 'm'}
	for _, r := range commonKeys {
		labels = append(labels, string(r))
		if len(labels) >= count {
			return labels
		}
	}

	// If we need more, use combinations of home row keys
	if count > len(labels) {
		for _, r1 := range homeRow {
			for _, r2 := range homeRow {
				labels = append(labels, string(r1)+string(r2))
				if len(labels) >= count {
					return labels
				}
			}
		}
	}

	return labels
}

// HandleJumpModeInput processes input during jump mode
func (b *Buffer) HandleJumpModeInput(ch rune) (jumped bool) {
	b.jumpInput += string(ch)

	// Check if we have a complete label
	if pos, ok := b.jumpPoints[b.jumpInput]; ok {
		b.cursorX = pos.X
		b.cursorY = pos.Y
		b.jumpMode = false
		b.jumpPoints = make(map[string]Position)
		b.jumpInput = ""
		return true
	}

	// Check if this input could still lead to a valid label
	for label := range b.jumpPoints {
		if strings.HasPrefix(label, b.jumpInput) {
			return false // Keep collecting input
		}
	}

	// No valid labels start with this input, reset jump mode
	b.ExitJumpMode()
	return false
}

// ExitJumpMode exits jump mode without jumping
func (b *Buffer) ExitJumpMode() {
	b.jumpMode = false
	b.jumpPoints = make(map[string]Position)
	b.jumpInput = ""
}

// GetJumpInput returns the current jump input
func (b *Buffer) GetJumpInput() string {
	return b.jumpInput
}

// IsJumpMode returns whether jump mode is active
func (b *Buffer) IsJumpMode() bool {
	return b.jumpMode
}

// GetJumpPoints returns current jump points
func (b *Buffer) GetJumpPoints() map[string]Position {
	return b.jumpPoints
}
