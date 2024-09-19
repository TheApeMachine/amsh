package components

import "bytes"

const content = `
 █████╗ ███╗   ███╗███████╗██╗  ██╗
██╔══██╗████╗ ████║██╔════╝██║  ██║
███████║██╔████╔██║███████╗███████║
██╔══██║██║╚██╔╝██║╚════██║██╔══██║
██║  ██║██║ ╚═╝ ██║███████║██║  ██║
╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝
The Ape Machine Shell v0.0.1
`

/*
Generator is used to construct the splash screen that is displayed when the shell is started.
It implements the io.Reader interface to allow it to be used in the render pipeline.
*/
type Generator struct {
	content *bytes.Buffer
}

/*
NewGenerator creates a new Generator with the content of the logo.
*/
func NewGenerator() *Generator {
	return &Generator{
		content: bytes.NewBufferString(content),
	}
}

/*
Read reads the content of the logo into the provided buffer.
*/
func (generator *Generator) Read(p []byte) (n int, err error) {
	return generator.content.Read(p)
}