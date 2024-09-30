package animation

import "time"

type Composite struct {
	Animations []Animation
	Duration   time.Duration
}

func (c Composite) Update(canvas *Canvas, progress float64) {
	for _, anim := range c.Animations {
		anim.Update(canvas, progress)
	}
}
