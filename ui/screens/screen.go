package screens

import (
	tea "github.com/charmbracelet/bubbletea"
)

type NextScreen tea.Msg

type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

func NewScreen(screenType Screen) Screen {
	return screenType
}
