package gopad

import (
	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const HelpOverlayID = "help"

var _ overlay.Overlay = (*HelpOverlay)(nil)

func NewHelpOverlay() HelpOverlay {
	h := config.NewHelp()

	return HelpOverlay{
		help: h,
	}
}

type HelpOverlay struct {
	help help.Model
}

func (h HelpOverlay) ID() string {
	return HelpOverlayID
}

func (h HelpOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (h HelpOverlay) Margin() (int, int) {
	return 0, 0
}

func (h HelpOverlay) Title() string {
	return "Help"
}

func (h HelpOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return h, nil
}

func (h HelpOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel, config.Keys.Help):
			cmds = append(cmds, overlay.Close(HelpOverlayID))
		}
	}

	return h, tea.Batch(cmds...)
}

func (h HelpOverlay) View(width int, height int) string {
	return h.help.View(config.Keys, width, height)
}
