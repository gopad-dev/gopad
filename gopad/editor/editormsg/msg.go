package editormsg

import (
	"github.com/charmbracelet/bubbletea/v2"
)

type Model int

const (
	ModelFile Model = iota
	ModelSearch
	ModelFileTree
)

func Focus(model Model) tea.Cmd {
	return func() tea.Msg {
		return FocusMsg{
			Model: model,
		}
	}
}

type FocusMsg struct {
	Model Model
}
