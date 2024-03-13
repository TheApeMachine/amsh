package mimic

import (
	"io"
	"os"

	"gopkg.in/qntfy/kazaam.v3"
)

type Shape struct {
	transformer *kazaam.Kazaam
}

func NewShape(spec string) *Shape {
	buf, _ := os.ReadFile("./cmd/cfg/" + spec)
	kz, _ := kazaam.NewKazaam(string(buf))
	return &Shape{kz}
}

func (shape *Shape) Read(p []byte) (n int, err error) {
	shape.transformer.TransformInPlace(p)
	return len(p), io.EOF
}
