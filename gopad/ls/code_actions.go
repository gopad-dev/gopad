package ls

import (
	tea "github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
)

func GetDeclaration(name string, row int, col int) tea.Cmd {
	return func() tea.Msg {
		return GetDeclarationMsg{
			Name: name,
			Row:  row,
			Col:  col,
		}
	}
}

type GetDeclarationMsg struct {
	Name string
	Row  int
	Col  int
}

func UpdateDeclaration(name string, declarations []Declaration) tea.Cmd {
	return func() tea.Msg {
		return UpdateDeclarationMsg{
			Name:         name,
			Declarations: declarations,
		}
	}
}

type UpdateDeclarationMsg struct {
	Name         string
	Declarations []Declaration
}

type Declaration struct {
	Name  string
	Range buffer.Range
}

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

func GetTypeDefinition(name string, row int, col int) tea.Cmd {
	return func() tea.Msg {
		return GetDefinitionMsg{
			Name: name,
			Row:  row,
			Col:  col,
		}
	}
}

type GetTypeDefinitionMsg struct {
	Name string
	Row  int
	Col  int
}

func UpdateTypeDefinition(name string, typeDefinitions []TypeDefinition) tea.Cmd {
	return func() tea.Msg {
		return UpdateTypeDefinitionMsg{
			Name:            name,
			TypeDefinitions: typeDefinitions,
		}
	}
}

type UpdateTypeDefinitionMsg struct {
	Name            string
	TypeDefinitions []TypeDefinition
}

type TypeDefinition struct {
	Name  string
	Range buffer.Range
}
