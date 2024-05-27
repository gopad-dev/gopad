package groupedlist

import (
	"github.com/charmbracelet/lipgloss"
)

var DefaultStyles = Styles{
	Style:                lipgloss.NewStyle().MarginLeft(1),
	ItemStyle:            lipgloss.NewStyle().Padding(0, 1),
	ItemSelectedStyle:    lipgloss.NewStyle().Padding(0, 1).Reverse(true),
	ItemDescriptionStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
}

type Styles struct {
	Style                lipgloss.Style
	ItemStyle            lipgloss.Style
	ItemSelectedStyle    lipgloss.Style
	ItemDescriptionStyle lipgloss.Style
}
