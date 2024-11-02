package core

type Buffer struct {
	Lines []string
}

// NewBuffer creates a new buffer with an initial empty line.
func NewBuffer() *Buffer {
	return &Buffer{
		Lines: []string{""},
	}
}

// Insert inserts a character at the specified line and column.
func (buffer *Buffer) Insert(line, col int, char rune) {
	if line >= len(buffer.Lines) {
		return
	}
	if col > len(buffer.Lines[line]) {
		col = len(buffer.Lines[line])
	}
	buffer.Lines[line] = buffer.Lines[line][:col] + string(char) + buffer.Lines[line][col:]
}

// Delete removes a character at the specified line and column.
func (buffer *Buffer) Delete(line, col int) {
	if line >= len(buffer.Lines) || col >= len(buffer.Lines[line]) {
		return
	}
	buffer.Lines[line] = buffer.Lines[line][:col] + buffer.Lines[line][col+1:]
}

// GetLine returns the content of a specific line.
func (buffer *Buffer) GetLine(line int) string {
	if line >= len(buffer.Lines) {
		return ""
	}
	return buffer.Lines[line]
}

// LineCount returns the number of lines in the buffer.
func (buffer *Buffer) LineCount() int {
	return len(buffer.Lines)
}

// InsertLine inserts a new empty line at the specified position
func (buffer *Buffer) InsertLine(at int) {
	if at < 0 || at > len(buffer.Lines) {
		return
	}
	buffer.Lines = append(buffer.Lines[:at], append([]string{""}, buffer.Lines[at:]...)...)
}

// SplitLine splits the line at the given position
func (buffer *Buffer) SplitLine(line, col int) {
	if line >= len(buffer.Lines) {
		return
	}
	currentLine := buffer.Lines[line]
	if col > len(currentLine) {
		col = len(currentLine)
	}

	// Split the line content
	newLine := currentLine[col:]
	buffer.Lines[line] = currentLine[:col]

	// Insert the new line
	buffer.Lines = append(buffer.Lines[:line+1], append([]string{newLine}, buffer.Lines[line+1:]...)...)
}
