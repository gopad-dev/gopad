package bubbles

import (
	"github.com/charmbracelet/bubbletea"
)

func IsKeyMsg(msg tea.Msg) bool {
	_, ok := msg.(tea.KeyMsg)
	return ok
}

func IsMouseMsg(msg tea.Msg) bool {
	_, ok := msg.(tea.MouseMsg)
	return ok
}

func IsInputMsg(msg tea.Msg) bool {
	if _, ok := msg.(tea.KeyMsg); ok {
		return true
	}

	_, ok := msg.(tea.MouseMsg)
	return ok
}
