package ls

import (
	tea "github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
)

func GetDefinition(name string, row int, col int) tea.Cmd {
	return func() tea.Msg {
		return GetDefinitionMsg{
			Name: name,
			Row:  row,
			Col:  col,
		}
	}
}

type GetDefinitionMsg struct {
	Name string
	Row  int
	Col  int
}

func UpdateDefinition(name string, definitions []Definition) tea.Cmd {
	return func() tea.Msg {
		return UpdateDefinitionMsg{
			Name:        name,
			Definitions: definitions,
		}
	}
}

type UpdateDefinitionMsg struct {
	Name        string
	Definitions []Definition
}

type Definition struct {
	Name  string
	Range buffer.Range
}
