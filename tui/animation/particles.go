package animation

import "math"

type Particles struct {
	Count int
}

func (p Particles) Update(canvas *Canvas, progress float64) {
	for i := 0; i < p.Count; i++ {
		angle := float64(i) * 2 * math.Pi / float64(p.Count)
		r := progress * float64(canvas.Width/4)
		x := int(float64(canvas.Width)/2 + r*math.Cos(angle))
		y := int(float64(canvas.Height)/2 + r*math.Sin(angle))
		canvas.Set(x, y, '*')
	}
}
