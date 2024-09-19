package core

/*
Layout is a structure that determines how multiple components are presented
within the screen space, as well as define how to most efficiently render them.
*/
type Layout struct {
	areas []*Area
}

// NewLayout initializes a new Layout with default settings
func NewLayout(areas ...*Area) Layout {
	return Layout{
		areas: areas,
	}
}

func (layout *Layout) Read(p []byte) (n int, err error) {
	for _, area := range layout.areas {
		if n, err = area.Read(p); err != nil {
			return
		}
	}
	return
}

func (layout *Layout) Write(p []byte) (n int, err error) {
	for _, area := range layout.areas {
		if n, err = area.Write(p); err != nil {
			return
		}
	}
	return
}

func (layout *Layout) Close() (err error) {
	for _, area := range layout.areas {
		if err = area.Close(); err != nil {
			return
		}
	}

	return
}
