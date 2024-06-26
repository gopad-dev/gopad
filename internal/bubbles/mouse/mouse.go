package mouse

import (
	"slices"

	"github.com/charmbracelet/bubbletea"
	"github.com/lrstanley/bubblezone"
)

func Matches(msg tea.MouseMsg, id string, button tea.MouseButton, actions ...tea.MouseAction) bool {
	z := zone.Get(id)

	if !z.InBounds(msg) {
		return false
	}

	if msg.Button != button {
		return false
	}

	if len(actions) == 0 {
		return true
	}

	return slices.Contains(actions, msg.Action)
}
