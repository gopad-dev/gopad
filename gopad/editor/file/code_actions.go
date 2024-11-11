package file

import (
	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/editor/buffer"

	"go.gopad.dev/gopad/gopad/ls"
)

func (f *Document) ShowDeclaration(p buffer.Point) tea.Cmd {
	return ls.GetDeclaration(f.Buffer.Name(), p)
}

func (f *Document) SetDeclarations(declarations []ls.FileLocation) {
	f.Declarations = declarations
}

func (f *Document) ShowDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetDefinition(f.Buffer.Name(), p)
}

func (f *Document) SetDefinitions(definitions []ls.FileLocation) {
	f.Definitions = definitions
}

func (f *Document) ShowTypeDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetTypeDefinition(f.Buffer.Name(), p)
}

func (f *Document) SetTypeDefinitions(typeDefinitions []ls.FileLocation) {
	f.TypeDefinitions = typeDefinitions
}

func (f *Document) ShowImplementations(p buffer.Point) tea.Cmd {
	return ls.GetImplementations(f.Buffer.Name(), p)
}

func (f *Document) SetImplementations(implementations []ls.FileLocation) {
	f.Implementations = implementations
}

func (f *Document) ShowReferences(p buffer.Point) tea.Cmd {
	return ls.GetReferences(f.Buffer.Name(), p)
}

func (f *Document) SetReferences(references []ls.FileLocation) {
	f.References = references
}
