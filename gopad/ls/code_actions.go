package ls

import (
	"github.com/charmbracelet/bubbletea/v2"
	"go.lsp.dev/protocol"

	"go.gopad.dev/gopad/internal/buffer"
)

func ParseLocations(locations []protocol.Location) []FileLocation {
	fileLocations := make([]FileLocation, len(locations))
	for i, location := range locations {
		fileLocations[i] = ParseLocation(location)
	}
	return fileLocations
}

func ParseLocation(location protocol.Location) FileLocation {
	return FileLocation{
		Name:  location.URI.Filename(),
		Range: buffer.ParseRange(location.Range),
	}
}

type FileLocation struct {
	Name  string
	Range buffer.Range
}

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

func UpdateDeclarations(name string, declarations []FileLocation) tea.Cmd {
	return func() tea.Msg {
		return UpdateDeclarationsMsg{
			Name:         name,
			Declarations: declarations,
		}
	}
}

type UpdateDeclarationsMsg struct {
	Name         string
	Declarations []FileLocation
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

func UpdateDefinitions(name string, definitions []FileLocation) tea.Msg {
	return UpdateDefinitionsMsg{
		Name:        name,
		Definitions: definitions,
	}
}

type UpdateDefinitionsMsg struct {
	Name        string
	Definitions []FileLocation
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

func UpdateTypeDefinitions(name string, typeDefinitions []FileLocation) tea.Cmd {
	return func() tea.Msg {
		return UpdateTypeDefinitionsMsg{
			Name:            name,
			TypeDefinitions: typeDefinitions,
		}
	}
}

type UpdateTypeDefinitionsMsg struct {
	Name            string
	TypeDefinitions []FileLocation
}

func GetImplementations(name string, p buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return GetImplementationsMsg{
			Name:  name,
			Point: p,
		}
	}
}

type GetImplementationsMsg struct {
	Name  string
	Point buffer.Point
}

func UpdateImplementations(name string, implementations []FileLocation) tea.Cmd {
	return func() tea.Msg {
		return UpdateImplementationsMsg{
			Name:            name,
			Implementations: implementations,
		}
	}
}

type UpdateImplementationsMsg struct {
	Name            string
	Implementations []FileLocation
}

func GetReferences(name string, p buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return GetReferencesMsg{
			Name:  name,
			Point: p,
		}
	}
}

type GetReferencesMsg struct {
	Name  string
	Point buffer.Point
}

func UpdateReferences(name string, references []FileLocation) tea.Cmd {
	return func() tea.Msg {
		return UpdateReferencesMsg{
			Name:       name,
			References: references,
		}
	}
}

type UpdateReferencesMsg struct {
	Name       string
	References []FileLocation
}
