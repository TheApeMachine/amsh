package types

/*
Mode determines the way the application behaves, especially when it comes to
key commands.
*/
type Mode uint

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
)
