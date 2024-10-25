package ls

import (
	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/internal/buffer"
)

func GetDeclaration(name string, p buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return GetDeclarationMsg{
			Name:  name,
			Point: p,
		}
	}
}

type GetDeclarationMsg struct {
	Name  string
	Point buffer.Point
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

func GetDefinition(name string, p buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return GetDefinitionMsg{
			Name:  name,
			Point: p,
		}
	}
}

type GetDefinitionMsg struct {
	Name  string
	Point buffer.Point
}

func UpdateDefinition(name string, definitions []Definition) tea.Msg {
	return UpdateDefinitionMsg{
		Name:        name,
		Definitions: definitions,
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

func GetTypeDefinition(name string, p buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return GetDefinitionMsg{
			Name:  name,
			Point: p,
		}
	}
}

type GetTypeDefinitionMsg struct {
	Name  string
	Point buffer.Point
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
