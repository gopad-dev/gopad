package gopad

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

func (s SetThemeOverlay) Init() tea.Cmd {
	return textinput.Blink
}

func (s SetThemeOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return s, overlay.Close(SetThemeOverlayID)
		case key.Matches(msg, config.Keys.OK):
			item := s.l.Selected()
			if item == nil {
				return s, nil
			}
			// TODO: set theme somehow
			// theme := item.(config.ThemeConfig)
			return s, tea.Batch(overlay.Close(SetThemeOverlayID))
		}
	}

	var cmd tea.Cmd
	s.l, cmd = s.l.Update(msg)

	return s, cmd
}

func (s SetThemeOverlay) View(width int, height int) string {
	s.l.SetHeight(height)
	s.l.SetWidth(width / 2)

	return s.l.View()
}
