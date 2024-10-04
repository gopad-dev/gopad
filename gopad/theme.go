package gopad

import (
	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

const SetThemeOverlayID = "theme"

var _ overlay.Overlay = (*SetThemeOverlay)(nil)

func NewSetThemeOverlay() SetThemeOverlay {
	l := config.NewList(config.Themes)
	l.TextInput.Placeholder = "Type a theme and press enter to set it"
	l.Focus()

	return SetThemeOverlay{
		l: l,
	}
}

type SetThemeOverlay struct {
	l list.Model[config.RawThemeConfig]
}

func (s SetThemeOverlay) ID() string {
	return SetThemeOverlayID
}

func (s SetThemeOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Top
}

func (s SetThemeOverlay) Margin() (int, int) {
	return 1, 1
}

func (s SetThemeOverlay) Title() string {
	return "Set Theme"
}

func (s SetThemeOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return s, textinput.Blink
}

func (s SetThemeOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return s, overlay.Close(SetThemeOverlayID)
		case key.Matches(msg, config.Keys.OK):
			theme := s.l.Selected()
			if theme.Name == "" {
				return s, nil
			}
			// TODO: set theme somehow
			// theme := item.(config.ThemeConfig)
			return s, tea.Batch(overlay.Close(SetThemeOverlayID))
		}
	}

	var cmd tea.Cmd
	s.l, cmd = s.l.Update(msg)

	if s.l.Clicked() {
		item := s.l.Selected()
		if item.Name != "" {
			// TODO: set theme somehow
			// theme := item.(config.ThemeConfig)
			return s, tea.Batch(cmd, overlay.Close(SetThemeOverlayID))
		}
	}

	return s, cmd
}

func (s SetThemeOverlay) View(width int, height int) string {
	s.l.SetHeight(height)
	s.l.SetWidth(width / 2)

	return s.l.View()
}
