package comms

type Proxy struct {
}

func NewProxy(parameters map[string]any) *Proxy {
	return &Proxy{}
}

func (proxy *Proxy) Start() string {
	return "Proxy started"
}
