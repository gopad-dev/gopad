package bubbles

import (
	tea "github.com/charmbracelet/bubbletea"
)

func IsInputMsg(msg tea.Msg) bool {
	if _, ok := msg.(tea.KeyMsg); ok {
		return true
	}

	if _, ok := msg.(tea.MouseDownMsg); ok {
		return true
	}

	if _, ok := msg.(tea.MouseUpMsg); ok {
		return true
	}

	if _, ok := msg.(tea.MouseWheelMsg); ok {
		return true
	}

	if _, ok := msg.(tea.MouseMotionMsg); ok {
		return true
	}

	return false
}
