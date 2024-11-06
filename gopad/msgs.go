package gopad

import (
	"github.com/charmbracelet/bubbletea/v2"
)

func OpenLSPOverlay() tea.Msg {
	return OpenLSPOverlayMsg{}
}

type OpenLSPOverlayMsg struct {
}
