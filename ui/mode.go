package ui

type Mode uint

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
)

func ModeFromString(modeStr string) Mode {
	switch modeStr {
	case "normal":
		return ModeNormal
	case "insert":
		return ModeInsert
	case "visual":
		return ModeVisual
	}

	return ModeNormal
}
