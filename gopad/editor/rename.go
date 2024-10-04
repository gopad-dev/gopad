package editor

import (
	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

const RenameOverlayID = "editor.rename"

var _ overlay.Overlay = (*RenameOverlay)(nil)

func NewRenameOverlay(name string) RenameOverlay {
	ti := config.NewTextInput()
	ti.Placeholder = "New file name"
	ti.SetValue(name)
	ti.Focus()

	return RenameOverlay{
		fileName: ti,
	}
}

type RenameOverlay struct {
	fileName textinput.Model
}

func (r RenameOverlay) ID() string {
	return RenameOverlayID
}

func (r RenameOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (r RenameOverlay) Margin() (int, int) {
	return 0, 0
}

func (r RenameOverlay) Title() string {
	return "Rename File"
}

func (r RenameOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return r, textinput.Blink
}

func (r RenameOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.OK):
			return r, tea.Sequence(
				overlay.Close(RenameOverlayID),
				file.RenameFile(r.fileName.Value()),
			)
		case key.Matches(msg, config.Keys.Cancel):
			return r, overlay.Close(RenameOverlayID)
		}
	}

	var cmd tea.Cmd
	r.fileName, cmd = r.fileName.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return r, tea.Batch(cmds...)
}

func (r RenameOverlay) View(width int, height int) string {
	return r.fileName.View()
}
