package gopad

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const KeyMapperOverlayID = "key.mapper"

var _ overlay.Overlay = (*KeyMapperOverlay)(nil)

func NewKeyMapperOverlay() KeyMapperOverlay {
	return KeyMapperOverlay{}
}

type KeyMapperOverlay struct {
	key string
}

func (q KeyMapperOverlay) ID() string {
	return KeyMapperOverlayID
}

func (q KeyMapperOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (q KeyMapperOverlay) Margin() (int, int) {
	return 0, 0
}

func (q KeyMapperOverlay) Title() string {
	return "Key Mapper"
}

func (q KeyMapperOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return q, nil
}

func (q KeyMapperOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return q, overlay.Close(KeyMapperOverlayID)
		default:
			q.key = msg.String()
			return q, nil
		}
	}

	return q, nil
}

func (q KeyMapperOverlay) View(width int, height int) string {
	if q.key == "" {
		return fmt.Sprintf("Press a key to see its name.\nPress %s to close.", config.Keys.Cancel.Help().Key)
	}
	return fmt.Sprintf("You pressed: \"%s\".\nPress %s to close.", q.key, config.Keys.Cancel.Help().Key)
}
