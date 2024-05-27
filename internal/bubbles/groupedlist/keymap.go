package groupedlist

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Start key.Binding
	End   key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "select previous entry"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "select next entry"),
	),
	Start: key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "select first entry"),
	),
	End: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "select last entry"),
	),
}
