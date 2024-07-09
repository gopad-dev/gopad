package bubbles

import (
	tea "github.com/charmbracelet/bubbletea"
)

func IsInputMsg(msg tea.Msg) bool {
	if _, ok := msg.(tea.KeyMsg); ok {
		return true
	}

	if _, ok := msg.(tea.MouseMsg); ok {
		return true
	}

	return false
}
