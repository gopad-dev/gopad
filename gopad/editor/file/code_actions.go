package file

import (
	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/buffer"
)

func (f *File) SetDeclarations(definitions []ls.Definition) tea.Cmd {
	if len(definitions) == 0 {
		return notifications.Add("No declaration found")
	}
	if len(definitions) == 1 {
		return f.openDefinition(definitions[0])
	}

	f.Definitions = definitions
	return nil
}

func (f *File) openDeclaration(definition ls.Definition) tea.Cmd {
	return OpenFilePosition(definition.Name, &buffer.Point{
		Row: definition.Range.Start.Row,
		Col: definition.Range.Start.Col,
	})
}

func (f *File) ShowDeclaration(p buffer.Point) tea.Cmd {
	return ls.GetDeclaration(f.Buffer.Name(), p)
}

func (f *File) SetDefinitions(definitions []ls.Definition) tea.Cmd {
	if len(definitions) == 0 {
		return notifications.Add("No definition found")
	}
	if len(definitions) == 1 {
		return f.openDefinition(definitions[0])
	}

	f.Definitions = definitions
	return nil
}

func (f *File) openDefinition(definition ls.Definition) tea.Cmd {
	return OpenFilePosition(definition.Name, &buffer.Point{
		Row: definition.Range.Start.Row,
		Col: definition.Range.Start.Col,
	})
}

func (f *File) ShowDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetDefinition(f.Buffer.Name(), p)
}

func (f *File) SetTypeDefinitions(typeDefinitions []ls.TypeDefinition) tea.Cmd {
	if len(typeDefinitions) == 0 {
		return notifications.Add("No type definition found")
	}
	if len(typeDefinitions) == 1 {
		return f.openTypeDefinition(typeDefinitions[0])
	}

	f.TypeDefinitions = typeDefinitions
	return nil
}

func (f *File) openTypeDefinition(typeDefinition ls.TypeDefinition) tea.Cmd {
	return OpenFilePosition(typeDefinition.Name, &buffer.Point{
		Row: typeDefinition.Range.Start.Row,
		Col: typeDefinition.Range.Start.Col,
	})
}

func (f *File) ShowTypeDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetTypeDefinition(f.Buffer.Name(), p)
}
