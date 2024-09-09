package filebrowser

import (
	"github.com/theapemachine/amsh/components"
)

func (model *Model) View() string {
	if model.state != components.Focused {
		return ""
	}

	return model.filepicker.View()
}
