package types

// Mode represents the current editor mode (Normal, Insert, etc.)
type Mode int

const (
	Normal Mode = iota
	Insert
	Command
	Visual
	BrowserMode
	ChatMode
)

// String returns a string representation of the mode
func (m Mode) String() string {
	switch m {
	case Normal:
		return "NORMAL"
	case Insert:
		return "INSERT"
	case Command:
		return "COMMAND"
	case Visual:
		return "VISUAL"
	case BrowserMode:
		return "BROWSER"
	case ChatMode:
		return "CHAT"
	default:
		return "UNKNOWN"
	}
}

// IsInput returns true if the mode accepts text input
func (m Mode) IsInput() bool {
	return m == Insert || m == Command || m == ChatMode
}

// IsMotion returns true if the mode accepts motion commands
func (m Mode) IsMotion() bool {
	return m == Normal || m == Visual || m == BrowserMode
}
