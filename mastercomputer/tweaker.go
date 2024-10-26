package mastercomputer

type Tweaker struct {
}

func NewTweaker(parameters map[string]any) *Tweaker {
	return &Tweaker{}
}

func (tweaker *Tweaker) Start() string {
	return "Tweaker started"
}
