package file

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

func (f *File) SetDefinitions(definitions []ls.Definition) tea.Cmd {
	log.Println("setting definitions", definitions)
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
	return OpenFilePosition(definition.Name, definition.Range.Start.Row, definition.Range.Start.Col)
}

func (f *File) ShowDefinitions() tea.Cmd {
	row, col := f.Cursor()

	return ls.GetDefinition(f.Name(), row, col)
}
