package animation

import "math"

type Sequence []Animation

func (s Sequence) Update(canvas *Canvas, progress float64) {
	if len(s) == 0 {
		return
	}
	index := int(progress * float64(len(s)))
	if index >= len(s) {
		index = len(s) - 1
	}
	s[index].Update(canvas, math.Mod(progress*float64(len(s)), 1.0))
}
