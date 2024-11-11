package file

import (
	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/editor/buffer"
	"go.gopad.dev/gopad/gopad/ls"
)

func (d *Document) ShowDeclaration(p buffer.Point) tea.Cmd {
	return ls.GetDeclaration(d.Name, p)
}

func (d *Document) SetDeclarations(declarations []ls.FileLocation) {
	d.Declarations = declarations
}

func (d *Document) ShowDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetDefinition(d.Name, p)
}

func (d *Document) SetDefinitions(definitions []ls.FileLocation) {
	d.Definitions = definitions
}

func (d *Document) ShowTypeDefinitions(p buffer.Point) tea.Cmd {
	return ls.GetTypeDefinition(d.Name, p)
}

func (d *Document) SetTypeDefinitions(typeDefinitions []ls.FileLocation) {
	d.TypeDefinitions = typeDefinitions
}

func (d *Document) ShowImplementations(p buffer.Point) tea.Cmd {
	return ls.GetImplementations(d.Name, p)
}

func (d *Document) SetImplementations(implementations []ls.FileLocation) {
	d.Implementations = implementations
}

func (d *Document) ShowReferences(p buffer.Point) tea.Cmd {
	return ls.GetReferences(d.Name, p)
}

func (d *Document) SetReferences(references []ls.FileLocation) {
	d.References = references
}
