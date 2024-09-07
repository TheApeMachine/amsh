package filebrowser

import (
	"strings"
)

func (model *Model) View() string {
	if !model.active {
		return ""
	}

	var s strings.Builder
	s.WriteString("\n  ")
	if model.err != nil {
		s.WriteString(model.filepicker.Styles.DisabledFile.Render(model.err.Error()))
	} else if model.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + model.filepicker.Styles.Selected.Render(model.selectedFile))
	}
	s.WriteString("\n\n" + model.filepicker.View() + "\n")
	return s.String()
}
