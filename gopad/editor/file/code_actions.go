package file

import (
	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/buffer"
)

func (f *File) ShowDeclaration(p buffer.Point) tea.Cmd {
	return ls.GetDeclaration(f.Buffer.Name(), p)
}

func (f *File) SetDeclarations(declarations []ls.FileLocation) {
	f.Declarations = declarations
}

func (f *File) ShowDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetDefinition(f.Buffer.Name(), p)
}

func (f *File) SetDefinitions(definitions []ls.FileLocation) {
	f.Definitions = definitions
}

func (f *File) ShowTypeDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetTypeDefinition(f.Buffer.Name(), p)
}

func (f *File) SetTypeDefinitions(typeDefinitions []ls.FileLocation) {
	f.TypeDefinitions = typeDefinitions
}

func (f *File) ShowImplementations(p buffer.Point) tea.Cmd {
	return ls.GetImplementations(f.Buffer.Name(), p)
}

func (f *File) SetImplementations(implementations []ls.FileLocation) {
	f.Implementations = implementations
}

func (f *File) ShowReferences(p buffer.Point) tea.Cmd {
	return ls.GetReferences(f.Buffer.Name(), p)
}

func (f *File) SetReferences(references []ls.FileLocation) {
	f.References = references
}
