package core

import "bytes"

type Buffer struct {
	render *bytes.Buffer
	cursor *Cursor
}

func NewBuffer() *Buffer {
	return &Buffer{}
}
