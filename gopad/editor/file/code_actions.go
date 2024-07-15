package file

import (
	"github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

func (f *File) SetDeclarations(definitions []ls.Definition) tea.Cmd {
	if len(definitions) == 0 {
		return notifications.Add("No declaration found")
	}
	if len(definitions) == 1 {
		return f.openDefinition(definitions[0])
	}

	f.definitions = definitions
	return nil
}

func (f *File) openDeclaration(definition ls.Definition) tea.Cmd {
	return OpenFilePosition(definition.Name, &buffer.Position{
		Row: definition.Range.Start.Row,
		Col: definition.Range.Start.Col,
	})
}

func (f *File) ShowDeclaration() tea.Cmd {
	row, col := f.Cursor()

	return ls.GetDeclaration(f.Name(), row, col)
}

func (f *File) SetDefinitions(definitions []ls.Definition) tea.Cmd {
	if len(definitions) == 0 {
		return notifications.Add("No definition found")
	}
	if len(definitions) == 1 {
		return f.openDefinition(definitions[0])
	}

	f.definitions = definitions
	return nil
}

func (f *File) openDefinition(definition ls.Definition) tea.Cmd {
	return OpenFilePosition(definition.Name, &buffer.Position{
		Row: definition.Range.Start.Row,
		Col: definition.Range.Start.Col,
	})
}

func (f *File) ShowDefinitions() tea.Cmd {
	row, col := f.Cursor()

	return ls.GetDefinition(f.Name(), row, col)
}

func (f *File) SetTypeDefinitions(typeDefinitions []ls.TypeDefinition) tea.Cmd {
	if len(typeDefinitions) == 0 {
		return notifications.Add("No type definition found")
	}
	if len(typeDefinitions) == 1 {
		return f.openTypeDefinition(typeDefinitions[0])
	}

	f.typeDefinitions = typeDefinitions
	return nil
}

func (f *File) openTypeDefinition(typeDefinition ls.TypeDefinition) tea.Cmd {
	return OpenFilePosition(typeDefinition.Name, &buffer.Position{
		Row: typeDefinition.Range.Start.Row,
		Col: typeDefinition.Range.Start.Col,
	})
}

func (f *File) ShowTypeDefinitions() tea.Cmd {
	row, col := f.Cursor()

	return ls.GetTypeDefinition(f.Name(), row, col)
}
