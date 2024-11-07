package core

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/theapemachine/amsh/errnie"
	"github.com/theapemachine/amsh/tui/features"
	"github.com/theapemachine/amsh/tui/types"
	boxer "github.com/treilik/bubbleboxer"
)

/*
Manager is the core manager of the application. It manages the screens and the
layout.
*/
type Manager struct {
	mode          types.Mode
	width, height int
	screens       boxer.Boxer
	screen        string
	keyBuffer     []string
	lastKeyTime   time.Time
}

func NewManager() *Manager {
	manager := &Manager{
		mode:    types.ModeNormal,
		screens: boxer.Boxer{},
		width:   80,
		height:  24,
	}

	manager.addScreens(true, features.NewSplash(manager.width, manager.height))
	return manager
}

/*
Init initializes the manager.
*/
func (manager *Manager) Init() tea.Cmd {
	// Collect all init commands from screens
	var cmds []tea.Cmd
	cmds = append(cmds, manager.screens.Init())

	// Add init commands from all models
	for _, model := range manager.screens.ModelMap {
		if cmd := model.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

/*
Update updates the manager.
*/
func (manager *Manager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle updates for all models, including animation ticks
	for _, model := range manager.screens.LayoutTree.Children {
		currentModel := manager.screens.ModelMap[model.GetAddress()]
		updatedModel, cmd := currentModel.Update(msg)

		// If we got an updated model, store it
		if updatedModel != nil {
			manager.screens.ModelMap[model.GetAddress()] = updatedModel
		}

		// Collect all commands
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		now := time.Now()

		if now.Sub(manager.lastKeyTime) > time.Second {
			manager.keyBuffer = []string{}
		}

		manager.lastKeyTime = now
		manager.keyBuffer = append(manager.keyBuffer, msg.String())

		if len(manager.keyBuffer) > 2 {
			manager.keyBuffer = manager.keyBuffer[1:]
		}

		if len(manager.keyBuffer) == 2 && manager.keyBuffer[0] == "q" && manager.keyBuffer[1] == "q" {
			return manager, tea.Quit
		}

		if len(manager.keyBuffer) == 2 && manager.keyBuffer[0] == " " && manager.keyBuffer[1] == "," {
			cmds = append(cmds, manager.addScreens(
				true,
				features.NewBrowser(),
				features.NewStatusBar(),
			))
		}

		if len(manager.keyBuffer) == 2 && manager.keyBuffer[0] == "c" && manager.keyBuffer[1] == "c" {
			for _, model := range manager.screens.ModelMap {
				if textarea, ok := model.(*features.TextArea); ok {
					highlightedText := textarea.GetHighlightedText()
					cmds = append(cmds, func() tea.Msg {
						return types.OpenChatMsg{Context: highlightedText}
					})
				}
			}
		}
	case features.FileSelectedMsg:
		// Switch out browser for textarea and statusbar
		cmds = append(cmds, manager.addScreens(
			true,
			features.NewTextarea(),
			features.NewStatusBar(),
		))

		// Forward the file path to the textarea
		cmds = append(cmds, func() tea.Msg {
			return features.LoadFileMsg{Filepath: msg.Path}
		})
	case tea.WindowSizeMsg:
		manager.width, manager.height = msg.Width, msg.Height

		// Update splash screen size if it exists
		for _, model := range manager.screens.ModelMap {
			if splash, ok := model.(*features.Splash); ok {
				splash.SetSize(msg.Width, msg.Height)
			}
		}

		errnie.MustVoid(manager.screens.UpdateSize(msg))
	}

	// Update the layout and collect any additional commands
	_, cmd := manager.screens.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return manager, tea.Batch(cmds...)
}

/*
View returns the manager's view.
*/
func (manager *Manager) View() string {
	return manager.screens.View()
}

/*
addScreens adds a new screen to the application.
*/
func (manager *Manager) addScreens(vertical bool, models ...features.Feature) tea.Cmd {
	var (
		nodes = []boxer.Node{}
		cmds  = []tea.Cmd{}
	)

	for _, model := range models {
		nodes = append(nodes, errnie.SafeMust(func() (boxer.Node, error) {
			return manager.screens.CreateLeaf(model.Name(), model)
		}))

		cmds = append(cmds, model.Init())
	}

	// Ensure minimum dimensions
	if manager.width < 10 {
		manager.width = 80
	}
	if manager.height < 10 {
		manager.height = 24
	}

	manager.screens.LayoutTree = boxer.Node{
		VerticalStacked: vertical,
		SizeFunc: func(node boxer.Node, widthOrHeight int) []int {
			sizes := make([]int, len(node.Children))
			remainingSpace := widthOrHeight

			// First pass: assign minimum sizes
			for i, child := range node.Children {
				if model, ok := manager.screens.ModelMap[child.GetAddress()]; ok {
					if feature, ok := model.(features.Feature); ok {
						width, height := feature.Size()
						if vertical {
							// Ensure minimum height of 3 for each component
							sizes[i] = max(height, 3)
						} else {
							// Ensure minimum width of 10 for each component
							sizes[i] = max(width, 10)
						}
						remainingSpace -= sizes[i]
					}
				}
			}

			// Adjust the sizes for vertical stacking
			if vertical && len(sizes) > 1 {
				// Assign remaining space to the top component
				sizes[0] += remainingSpace
			}

			return sizes
		},
		Children: nodes,
	}

	errnie.MustVoid(manager.screens.UpdateSize(
		tea.WindowSizeMsg{Width: manager.width, Height: manager.height},
	))

	manager.screen = models[0].Name()

	return tea.Batch(cmds...)
}
