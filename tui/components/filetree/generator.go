package filetree

import (
	"bytes"
)

type Generator struct {
	content *bytes.Buffer
}

func NewGenerator() *Generator {
	return &Generator{
		content: bytes.NewBuffer([]byte{}),
	}
}

func (generator *Generator) Read(p []byte) (n int, err error) {
	return generator.content.Read(p)
}

func (generator *Generator) Write(p []byte) (n int, err error) {
	return generator.content.Write(p)
}

func (generator *Generator) Close() error {
	return nil
}
