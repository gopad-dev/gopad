package editor

import (
	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

const SetLanguageOverlayID = "file.language"
const FileLanguageZoneID = "file.language"

var _ overlay.Overlay = (*SetLanguageOverlay)(nil)

func NewSetLanguageOverlay() SetLanguageOverlay {
	l := config.NewList(file.Languages)
	l.TextInput.Placeholder = "Type a language and press enter to set it"
	l.Focus()

	return SetLanguageOverlay{
		l: l,
	}
}

type SetLanguageOverlay struct {
	l list.Model[*file.Language]
}

func (s SetLanguageOverlay) ID() string {
	return SetLanguageOverlayID
}

func (s SetLanguageOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Top
}

func (s SetLanguageOverlay) Margin() (int, int) {
	return 1, 1
}

func (s SetLanguageOverlay) Title() string {
	return "Set Language"
}

func (s SetLanguageOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return s, textinput.Blink
}

func (s SetLanguageOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return s, overlay.Close(SetLanguageOverlayID)
		case key.Matches(msg, config.Keys.OK):
			lang := s.l.Selected()
			if lang == nil {
				return s, nil
			}
			return s, tea.Batch(overlay.Close(SetLanguageOverlayID), file.SetLanguage(lang.Name))
		}
	}

	var cmd tea.Cmd
	s.l, cmd = s.l.Update(msg)

	if s.l.Clicked() {
		lang := s.l.Selected()
		return s, tea.Batch(cmd, overlay.Close(SetLanguageOverlayID), file.SetLanguage(lang.Name))
	}

	return s, cmd
}

func (s SetLanguageOverlay) View(width int, height int) string {
	s.l.SetHeight(height)
	s.l.SetWidth(width / 2)

	return s.l.View()
}
