package core

import "io"

/*
Area is a structure that represents a rectangular area on the screen.
*/
type Area struct {
	component     io.ReadWriteCloser
	x, y          int
	width, height int
}

/*
NewArea creates a new area with the given dimensions.
*/
func NewArea(component io.ReadWriteCloser, x, y, width, height int) *Area {
	return &Area{
		component: component,
		x:         x,
		y:         y,
		width:     width,
		height:    height,
	}
}

func (area *Area) Read(p []byte) (n int, err error) {
	return area.component.Read(p)
}

func (area *Area) Write(p []byte) (n int, err error) {
	return area.component.Write(p)
}

func (area *Area) Close() error {
	return area.component.Close()
}
