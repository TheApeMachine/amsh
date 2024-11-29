package textarea

/*
FROM: https://github.com/charmbracelet/bubbles/blob/master/textarea/textarea.go
*/

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea/memoization"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	minHeight        = 1
	defaultHeight    = 6
	defaultWidth     = 40
	defaultCharLimit = 400
	defaultMaxHeight = 99
	defaultMaxWidth  = 500
)

// KeyMap is the key bindings for different actions within the textarea.
type KeyMap struct {
	CharacterBackward       key.Binding
	CharacterForward        key.Binding
	DeleteAfterCursor       key.Binding
	DeleteBeforeCursor      key.Binding
	DeleteCharacterBackward key.Binding
	DeleteCharacterForward  key.Binding
	DeleteWordBackward      key.Binding
	DeleteWordForward       key.Binding
	InsertNewline           key.Binding
	LineEnd                 key.Binding
	LineNext                key.Binding
	LinePrevious            key.Binding
	LineStart               key.Binding
	Paste                   key.Binding
	WordBackward            key.Binding
	WordForward             key.Binding
	InputBegin              key.Binding
	InputEnd                key.Binding

	UppercaseWordForward  key.Binding
	LowercaseWordForward  key.Binding
	CapitalizeWordForward key.Binding

	TransposeCharacterBackward key.Binding
}

// DefaultKeyMap is the default set of key bindings for navigating and acting
// upon the textarea.
var DefaultKeyMap = KeyMap{
	CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f"), key.WithHelp("right", "character forward")),
	CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b"), key.WithHelp("left", "character backward")),
	WordForward:             key.NewBinding(key.WithKeys("alt+right", "alt+f"), key.WithHelp("alt+right", "word forward")),
	WordBackward:            key.NewBinding(key.WithKeys("alt+left", "alt+b"), key.WithHelp("alt+left", "word backward")),
	LineNext:                key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down", "next line")),
	LinePrevious:            key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up", "previous line")),
	DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w"), key.WithHelp("alt+backspace", "delete word backward")),
	DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d"), key.WithHelp("alt+delete", "delete word forward")),
	DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "delete after cursor")),
	DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "delete before cursor")),
	InsertNewline:           key.NewBinding(key.WithKeys("enter", "ctrl+m"), key.WithHelp("enter", "insert newline")),
	DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h"), key.WithHelp("backspace", "delete character backward")),
	DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d"), key.WithHelp("delete", "delete character forward")),
	LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a"), key.WithHelp("home", "line start")),
	LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e"), key.WithHelp("end", "line end")),
	Paste:                   key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste")),
	InputBegin:              key.NewBinding(key.WithKeys("alt+<", "ctrl+home"), key.WithHelp("alt+<", "input begin")),
	InputEnd:                key.NewBinding(key.WithKeys("alt+>", "ctrl+end"), key.WithHelp("alt+>", "input end")),

	CapitalizeWordForward: key.NewBinding(key.WithKeys("alt+c"), key.WithHelp("alt+c", "capitalize word forward")),
	LowercaseWordForward:  key.NewBinding(key.WithKeys("alt+l"), key.WithHelp("alt+l", "lowercase word forward")),
	UppercaseWordForward:  key.NewBinding(key.WithKeys("alt+u"), key.WithHelp("alt+u", "uppercase word forward")),

	TransposeCharacterBackward: key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "transpose character backward")),
}

// Region represents a styled section of text
type Region struct {
	ID              string
	ForegroundColor string
	BackgroundColor string
	Attributes      string
	StartPos        int
	EndPos          int
}

// LineInfo holds information about a line's regions and display properties
type LineInfo struct {
	// Width is the number of columns in the line.
	Width int
	// CharWidth is the number of characters in the line to account for
	// double-width runes.
	CharWidth int
	// Height is the number of rows in the line.
	Height int
	// StartColumn is the index of the first column of the line.
	StartColumn int
	// ColumnOffset is the number of columns that the cursor is offset from the
	// start of the line.
	ColumnOffset int
	// RowOffset is the number of rows that the cursor is offset from the start
	// of the line.
	RowOffset int
	// CharOffset is the number of characters that the cursor is offset
	// from the start of the line. This will generally be equivalent to
	// ColumnOffset, but will be different there are double-width runes before
	// the cursor.
	CharOffset int

	// Region-based styling support
	Regions    []Region
	DisplayLen int
	StartPos   int
}

// Style that will be applied to the text area.
//
// Style can be applied to focused and unfocused states to change the styles
// depending on the focus state.
//
// For an introduction to styling with Lip Gloss see:
// https://github.com/charmbracelet/lipgloss
type Style struct {
	Base             lipgloss.Style
	CursorLine       lipgloss.Style
	CursorLineNumber lipgloss.Style
	EndOfBuffer      lipgloss.Style
	LineNumber       lipgloss.Style
	Placeholder      lipgloss.Style
	Prompt           lipgloss.Style
	Text             lipgloss.Style
}

func (s Style) computedCursorLine() lipgloss.Style {
	return s.CursorLine.Inherit(s.Base).Inline(true)
}

func (s Style) computedCursorLineNumber() lipgloss.Style {
	return s.CursorLineNumber.
		Inherit(s.CursorLine).
		Inherit(s.Base).
		Inline(true)
}

func (s Style) computedEndOfBuffer() lipgloss.Style {
	return s.EndOfBuffer.Inherit(s.Base).Inline(true)
}

func (s Style) computedLineNumber() lipgloss.Style {
	return s.LineNumber.Inherit(s.Base).Inline(true)
}

func (s Style) computedPrompt() lipgloss.Style {
	return s.Prompt.Inherit(s.Base).Inline(true)
}

func (s Style) computedText() lipgloss.Style {
	return s.Text.Inherit(s.Base).Inline(true)
}

func (s Style) computedPlaceholder() lipgloss.Style {
	return s.Placeholder.Inherit(s.Base).Inline(true)
}

// line is the input to the text wrapping function. This is stored in a struct
// so that it can be hashed and memoized.
type line struct {
	runes []rune
	width int
}

// Hash returns a hash of the line.
func (w line) Hash() string {
	v := fmt.Sprintf("%s:%d", string(w.runes), w.width)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v)))
}

// Model is the Bubble Tea model for this text area element
type Model struct {
	Err error

	// General settings
	cache *memoization.MemoCache[line, [][]rune]

	// Buffer holds the text content
	buffer *Buffer

	// Prompt is printed at the beginning of each line
	Prompt string

	// Placeholder is the text displayed when the user hasn't entered anything yet
	Placeholder string

	// ShowLineNumbers, if enabled, causes line numbers to be printed after the prompt
	ShowLineNumbers bool

	// EndOfBufferCharacter is displayed at the end of the input
	EndOfBufferCharacter rune

	// KeyMap encodes the keybindings recognized by the widget
	KeyMap KeyMap

	// Styling
	FocusedStyle Style
	BlurredStyle Style
	style        *Style

	// Cursor is the text area cursor
	Cursor cursor.Model

	// CharLimit is the maximum number of characters this input element will accept
	CharLimit int

	// MaxHeight is the maximum height of the text area in rows
	MaxHeight int

	// MaxWidth is the maximum width of the text area in columns
	MaxWidth int

	// If promptFunc is set, it replaces Prompt as a generator for prompt strings
	promptFunc func(line int) string

	// width is the maximum width of the text area
	width int

	// height is the maximum height of the text area
	height int

	// viewport is the scroll viewport
	viewport viewport.Model

	// row is the cursor row
	row int

	// col is the cursor column in the buffer
	col int

	// focused is true when the textarea is focused
	focused bool

	blockInput        bool
	selecting         bool
	selectionStart    int
	selectionStartCol int
}

// New creates a new text area.
func New() Model {
	m := Model{
		buffer:               NewBuffer(),
		cache:                memoization.NewMemoCache[line, [][]rune](200),
		style:                &DefaultStyle,
		Cursor:               cursor.New(),
		CharLimit:            defaultCharLimit,
		MaxHeight:            defaultMaxHeight,
		MaxWidth:             defaultMaxWidth,
		EndOfBufferCharacter: '~',
		viewport:             viewport.New(0, 0),
		blockInput:           false,
		selecting:            false,
		selectionStart:       0,
		selectionStartCol:    0,
		focused:              false,
	}

	m.style = &m.BlurredStyle
	return m
}

// SetContent sets the content of the text area.
func (m *Model) SetContent(s string) {
	m.buffer = NewBuffer()
	m.buffer.Insert(0, 0, s)
	m.row = 0
	m.col = 0
}

// Content returns the content of the text area.
func (m Model) Content() string {
	var s strings.Builder
	for i, line := range m.buffer.lines {
		if i > 0 {
			s.WriteRune('\n')
		}
		s.WriteString(string(line))
	}
	return s.String()
}

// Value returns the value of the text area.
func (m Model) Value() string {
	return m.Content()
}

// LineCount returns the number of lines in the text area.
func (m Model) LineCount() int {
	return len(m.buffer.lines)
}

// Line returns the line at the given index.
func (m Model) Line(row int) string {
	if row >= len(m.buffer.lines) {
		return ""
	}
	return string(m.buffer.lines[row])
}

// LineInfo returns information about the cursor line.
func (m Model) LineInfo() LineInfo {
	displayRow, displayCol := m.buffer.GetDisplayPosition(m.row, m.col)
	return LineInfo{
		Width:        m.width,
		CharWidth:    displayCol,
		ColumnOffset: displayCol,
		RowOffset:    displayRow,
		StartColumn:  0,
		Height:       1,
	}
}

// Internal messages for clipboard operations
type (
	pasteMsg    string
	pasteErrMsg struct{ error }
)

// Update method should handle paste operations
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pasteMsg:
		if !m.blockInput {
			m.buffer.Insert(m.row, m.col, string(msg))
			m.col += len(msg)
		}
	case pasteErrMsg:
		m.Err = msg
	case tea.KeyMsg:
		// Allow navigation keys even when input is blocked
		switch msg.Type {
		case tea.KeyLeft:
			if m.col > 0 {
				m.col--
			} else if m.row > 0 {
				m.row--
				m.col = len(m.buffer.lines[m.row])
			}
		case tea.KeyRight:
			if m.col < len(m.buffer.lines[m.row]) {
				m.col++
			} else if m.row < len(m.buffer.lines)-1 {
				m.row++
				m.col = 0
			}
		case tea.KeyUp:
			if m.row > 0 {
				m.row--
				if m.col > len(m.buffer.lines[m.row]) {
					m.col = len(m.buffer.lines[m.row])
				}
			}
		case tea.KeyDown:
			if m.row < len(m.buffer.lines)-1 {
				m.row++
				if m.col > len(m.buffer.lines[m.row]) {
					m.col = len(m.buffer.lines[m.row])
				}
			}
		default:
			// Only block non-navigation keys
			if !m.blockInput {
				switch msg.Type {
				case tea.KeyEnter:
					m.buffer.Insert(m.row, m.col, "\n")
					m.row++
					m.col = 0
				case tea.KeyBackspace:
					if m.col > 0 {
						m.buffer.Delete(m.row, m.col-1, 1)
						m.col--
					} else if m.row > 0 {
						prevLineLen := len(m.buffer.lines[m.row-1])
						m.buffer.Delete(m.row-1, prevLineLen, 1)
						m.row--
						m.col = prevLineLen
					}
				case tea.KeyRunes:
					m.buffer.Insert(m.row, m.col, string(msg.Runes))
					m.col += len(msg.Runes)
				}
			}
		}
	}

	// Update viewport
	vp, cmd := m.viewport.Update(msg)
	m.viewport = vp
	cmds = append(cmds, cmd)

	// Update cursor
	cur, cmd := m.Cursor.Update(msg)
	m.Cursor = cur
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Paste is a command for pasting from the clipboard into the text input
func Paste() tea.Msg {
	str, err := clipboard.ReadAll()
	if err != nil {
		return pasteErrMsg{err}
	}
	return pasteMsg(str)
}

// View renders the textarea
func (m Model) View() string {
	// Hard limit the viewport width
	m.viewport.Width = min(214, m.width)
	m.viewport.Height = m.height

	var lines []string
	for row, line := range m.buffer.lines {
		var sb strings.Builder

		// Add prompt
		prompt := m.style.computedPrompt().Render(m.getPromptString(row))
		sb.WriteString(prompt)

		// Add line numbers if enabled
		if m.ShowLineNumbers {
			lineNum := m.formatLineNumber(row + 1)
			if row == m.row {
				sb.WriteString(m.style.computedCursorLineNumber().Render(lineNum))
			} else {
				sb.WriteString(m.style.computedLineNumber().Render(lineNum))
			}
		}

		// Calculate remaining width
		usedWidth := lipgloss.Width(sb.String())
		remainingWidth := m.viewport.Width - usedWidth

		if remainingWidth <= 0 {
			lines = append(lines, sb.String())
			continue
		}

		// Get line info and regions
		if m.buffer.modified {
			m.buffer.reindex()
		}
		lineInfo := m.buffer.lineInfo[row]

		// Handle cursor line
		if row == m.row {
			// Split content at cursor
			cursorPos := min(m.col, len(line))

			// Find region at cursor
			var cursorRegion *Region
			for i := range lineInfo.Regions {
				if cursorPos >= lineInfo.Regions[i].StartPos && cursorPos < lineInfo.Regions[i].EndPos {
					cursorRegion = &lineInfo.Regions[i]
					break
				}
			}

			// Render content with cursor
			if cursorRegion != nil {
				// Render region content up to cursor
				before := string(line[cursorRegion.StartPos:cursorPos])
				after := string(line[cursorPos:cursorRegion.EndPos])

				style := m.style.computedCursorLine()
				if cursorRegion.ForegroundColor != "" {
					style = style.Foreground(lipgloss.Color(cursorRegion.ForegroundColor))
				}

				sb.WriteString(style.Render(before))
				m.Cursor.SetChar(" ")
				sb.WriteString(m.Cursor.View())
				sb.WriteString(style.Render(after))
			} else {
				// No region at cursor
				sb.WriteString(m.style.computedCursorLine().Render(string(line[:cursorPos])))
				m.Cursor.SetChar(" ")
				sb.WriteString(m.Cursor.View())
				if cursorPos < len(line) {
					sb.WriteString(m.style.computedCursorLine().Render(string(line[cursorPos:])))
				}
			}
		} else {
			// Render regular line with regions
			for _, region := range lineInfo.Regions {
				content := string(line[region.StartPos:region.EndPos])
				style := m.style.computedText()
				if region.ForegroundColor != "" {
					style = style.Foreground(lipgloss.Color(region.ForegroundColor))
				}
				sb.WriteString(style.Render(content))
			}
		}

		lines = append(lines, sb.String())
	}

	// Fill remaining space
	for row := len(m.buffer.lines); row < m.height; row++ {
		var sb strings.Builder
		sb.WriteString(m.style.computedPrompt().Render(m.getPromptString(row)))
		if m.ShowLineNumbers {
			sb.WriteString(m.style.computedLineNumber().Render(m.formatLineNumber(" ")))
		}
		sb.WriteString(m.style.computedEndOfBuffer().Render(string(m.EndOfBufferCharacter)))
		lines = append(lines, sb.String())
	}

	// Set content with width enforcement
	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)
	return m.style.Base.MaxWidth(m.viewport.Width).Render(m.viewport.View())
}

// Helper functions
func (m Model) formatLineNumber(x any) string {
	digits := len(strconv.Itoa(m.MaxHeight))
	return fmt.Sprintf(" %*v ", digits, x)
}

func (m Model) getPromptString(displayLine int) string {
	if m.promptFunc != nil {
		return m.promptFunc(displayLine)
	}
	return m.Prompt
}

// Width returns the width of the textarea
func (m Model) Width() int {
	return m.width
}

// Height returns the height of the textarea
func (m Model) Height() int {
	return m.height
}

// SetWidth sets the width of the textarea
func (m *Model) SetWidth(width int) {
	m.width = width
	m.viewport.Width = width
}

// SetHeight sets the height of the textarea
func (m *Model) SetHeight(height int) {
	m.height = height
	m.viewport.Height = height
}

// SetValue sets the content of the textarea
func (m *Model) SetValue(s string) {
	m.buffer = NewBuffer()
	m.buffer.Insert(0, 0, s)
	m.row = 0
	m.col = 0
}

// InsertString inserts a string at the current cursor position
func (m *Model) InsertString(s string) {
	m.buffer.Insert(m.row, m.col, s)
	m.col += len(s)
}

// CursorStart moves the cursor to the start of the text
func (m *Model) CursorStart() {
	m.row = 0
	m.col = 0
}

// Focus gives focus to the textarea
func (m *Model) Focus() tea.Cmd {
	m.focused = true
	m.style = &m.FocusedStyle
	return m.Cursor.Focus()
}

// Blur removes focus from the textarea
func (m *Model) Blur() {
	m.focused = false
	m.style = &m.BlurredStyle
	m.Cursor.Blur()
}

// SetBlockInput sets whether input is blocked
func (m *Model) SetBlockInput(block bool) {
	m.blockInput = true
}

// StartSelection starts text selection
func (m *Model) StartSelection() {
	m.selecting = true
	m.selectionStart = m.row
	m.selectionStartCol = m.col
}

// EndSelection ends text selection
func (m *Model) EndSelection() {
	m.selecting = false
}

// Blink returns the cursor blink command
func Blink() tea.Msg {
	return cursor.Blink()
}

// DefaultStyle is the default style for the textarea
var DefaultStyle = Style{
	Base:             lipgloss.NewStyle(),
	CursorLine:       lipgloss.NewStyle().Background(lipgloss.Color("62")),
	CursorLineNumber: lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
	EndOfBuffer:      lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	LineNumber:       lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
	Prompt:           lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
	Text:             lipgloss.NewStyle(),
}

// Buffer manages text content with style information
type Buffer struct {
	lines    [][]rune   // Raw text including ANSI codes
	lineInfo []LineInfo // Processed line information
	modified bool       // Whether buffer needs reindexing
}

// NewBuffer creates a new text buffer
func NewBuffer() *Buffer {
	return &Buffer{
		lines:    make([][]rune, 1),
		lineInfo: make([]LineInfo, 1),
	}
}

// processANSICodes extracts ANSI codes and creates regions
func (b *Buffer) processANSICodes(line []rune) LineInfo {
	info := LineInfo{
		Regions: make([]Region, 0),
	}

	var currentRegion Region
	inANSI := false
	var ansiCode strings.Builder
	displayLen := 0
	pos := 0
	charWidth := 0
	columnOffset := 0

	for i, r := range line {
		if r == '\x1b' {
			inANSI = true
			if currentRegion.StartPos < i {
				currentRegion.EndPos = i
				info.Regions = append(info.Regions, currentRegion)
			}
			ansiCode.Reset()
			ansiCode.WriteRune(r)
			continue
		}

		if inANSI {
			ansiCode.WriteRune(r)
			if r == 'm' {
				// Process ANSI code
				code := ansiCode.String()
				currentRegion = Region{
					ID:              fmt.Sprintf("ansi-%d", len(info.Regions)),
					ForegroundColor: code,
					StartPos:        i + 1,
				}
				inANSI = false
			}
			continue
		}

		width := runewidth.RuneWidth(r)
		displayLen += width
		columnOffset += width
		charWidth++
		pos = i
	}

	if currentRegion.StartPos < pos {
		currentRegion.EndPos = pos + 1
		info.Regions = append(info.Regions, currentRegion)
	}

	info.DisplayLen = displayLen
	info.Width = displayLen
	info.CharWidth = charWidth
	info.ColumnOffset = columnOffset
	info.Height = 1
	return info
}

// reindex rebuilds the line information
func (b *Buffer) reindex() {
	b.lineInfo = make([]LineInfo, len(b.lines))
	for i, line := range b.lines {
		b.lineInfo[i] = b.processANSICodes(line)
	}
	b.modified = false
}

// Insert adds text at the specified position
func (b *Buffer) Insert(row, col int, text string) {
	// Ensure we have enough lines
	for len(b.lines) <= row {
		b.lines = append(b.lines, make([]rune, 0))
		b.lineInfo = append(b.lineInfo, LineInfo{})
	}

	// Convert text to runes
	runes := []rune(text)

	// Handle newlines in the inserted text
	if strings.Contains(text, "\n") {
		lines := strings.Split(text, "\n")
		// Handle first line (insert into current line)
		firstLine := []rune(lines[0])
		b.lines[row] = append(b.lines[row][:col], append(firstLine, b.lines[row][col:]...)...)

		// Insert remaining lines
		newLines := make([][]rune, 0, len(lines)-1)
		for _, line := range lines[1:] {
			newLines = append(newLines, []rune(line))
		}

		// Splice in the new lines
		b.lines = append(b.lines[:row+1], append(newLines, b.lines[row+1:]...)...)
	} else {
		// Simple insertion within a line
		b.lines[row] = append(b.lines[row][:col], append(runes, b.lines[row][col:]...)...)
	}

	b.modified = true
}

// Delete removes text at the specified position
func (b *Buffer) Delete(row, col, count int) {
	if row >= len(b.lines) {
		return
	}

	line := b.lines[row]
	if col >= len(line) {
		return
	}

	end := col + count
	if end > len(line) {
		end = len(line)
	}

	b.lines[row] = append(line[:col], line[end:]...)
	b.modified = true
}

// GetDisplayPosition returns the display coordinates for a given buffer position
func (b *Buffer) GetDisplayPosition(row, col int) (displayRow, displayCol int) {
	if row >= len(b.lineInfo) {
		return 0, 0
	}

	// Calculate display column by counting visible width up to col
	displayCol = 0
	inANSI := false
	for i := 0; i < col && i < len(b.lines[row]); i++ {
		r := b.lines[row][i]
		if r == '\x1b' {
			inANSI = true
			continue
		}
		if inANSI {
			if r == 'm' {
				inANSI = false
			}
			continue
		}
		displayCol += runewidth.RuneWidth(r)
	}

	return row, displayCol
}

// GetBufferPosition converts a display position to a buffer position
func (b *Buffer) GetBufferPosition(displayRow, displayCol int) (row, col int) {
	if b.modified {
		b.reindex()
	}

	if displayRow >= len(b.lineInfo) {
		return 0, 0
	}

	row = displayRow
	col = 0
	currentDisplayCol := 0
	inANSI := false
	var currentANSI strings.Builder

	// Process each character until we reach the target display column
	for i, r := range b.lines[row] {
		if r == '\x1b' {
			inANSI = true
			currentANSI.WriteRune(r)
			continue
		}

		if inANSI {
			currentANSI.WriteRune(r)
			if r == 'm' {
				inANSI = false
				currentANSI.Reset()
			}
			continue
		}

		width := runewidth.RuneWidth(r)
		if currentDisplayCol+width > displayCol {
			break
		}
		currentDisplayCol += width
		col = i + 1
	}

	return row, col
}
