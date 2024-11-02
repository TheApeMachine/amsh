package core

type Cursor struct {
	line int
	col  int
}

// NewCursor creates a new cursor at the start position.
func NewCursor() *Cursor {
	return &Cursor{line: 0, col: 0}
}

// MoveUp moves the cursor up by one line.
func (cursor *Cursor) MoveUp() {
	if cursor.line > 0 {
		cursor.line--
	}
}

// MoveDown moves the cursor down by one line.
func (cursor *Cursor) MoveDown(buffer *Buffer) {
	if cursor.line < buffer.LineCount()-1 {
		cursor.line++
	}
}

// MoveLeft moves the cursor left by one column.
func (cursor *Cursor) MoveLeft() {
	if cursor.col > 0 {
		cursor.col--
	}
}

// MoveRight moves the cursor right by one column.
func (cursor *Cursor) MoveRight(buffer *Buffer) {
	if cursor.col < len(buffer.GetLine(cursor.line)) {
		cursor.col++
	}
}

// GetPosition returns the current line and column of the cursor.
func (cursor *Cursor) GetPosition() (int, int) {
	return cursor.line, cursor.col
}

// MoveTo moves the cursor to a specific position
func (cursor *Cursor) MoveTo(line, col int) {
	cursor.line = line
	cursor.col = col
}
