package editor

type Mode int

const (
	NormalMode Mode = iota
	InsertMode
)

func (m Mode) String() string {
	switch m {
	case NormalMode:
		return "NORMAL"
	case InsertMode:
		return "INSERT"
	default:
		return "UNKNOWN"
	}
}