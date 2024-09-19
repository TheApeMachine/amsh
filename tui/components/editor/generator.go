package editor

import (
	"bufio"
	"bytes"
)

type Generator struct {
	buf     *bytes.Buffer
	reader  *bufio.Reader
	writer  *bufio.Writer
	lines   []string
	cursorX int
	cursorY int
}

func NewGenerator(width, height int) *Generator {
	buf := bytes.NewBuffer([]byte{})
	return &Generator{
		buf:     buf,
		reader:  bufio.NewReader(buf),
		writer:  bufio.NewWriter(buf),
		lines:   make([]string, height),
		cursorX: 0,
		cursorY: 0,
	}
}

func (generator *Generator) Initialize() {
	generator.lines = []string{
		"Hello, World!",
		"This is a simple TUI",
		"Use 'w', 'a', 's', 'd' to move.",
		"Press 'q' to quit.",
	}
}

func (generator *Generator) Read(p []byte) (n int, err error) {
	return generator.reader.Read(p)
}

func (generator *Generator) Write(p []byte) (n int, err error) {
	if n, err = generator.writer.Write(p); err == nil {
		err = generator.writer.Flush()
	}

	return n, err
}

func (generator *Generator) Close() error {
	return generator.writer.Flush()
}
