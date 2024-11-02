package core

type Buffer struct {
	lines []string
}

// NewBuffer creates a new buffer with an initial empty line.
func NewBuffer() *Buffer {
	return &Buffer{
		lines: []string{""},
	}
}

// Insert inserts a character at the specified line and column.
func (buffer *Buffer) Insert(line, col int, char rune) {
	if line >= len(buffer.lines) {
		return
	}
	if col > len(buffer.lines[line]) {
		col = len(buffer.lines[line])
	}
	buffer.lines[line] = buffer.lines[line][:col] + string(char) + buffer.lines[line][col:]
}

// Delete removes a character at the specified line and column.
func (buffer *Buffer) Delete(line, col int) {
	if line >= len(buffer.lines) || col >= len(buffer.lines[line]) {
		return
	}
	buffer.lines[line] = buffer.lines[line][:col] + buffer.lines[line][col+1:]
}

// GetLine returns the content of a specific line.
func (buffer *Buffer) GetLine(line int) string {
	if line >= len(buffer.lines) {
		return ""
	}
	return buffer.lines[line]
}

// LineCount returns the number of lines in the buffer.
func (buffer *Buffer) LineCount() int {
	return len(buffer.lines)
}
