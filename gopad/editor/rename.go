package editor

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

func (h RenameOverlay) ID() string {
	return RenameOverlayID
}

func (h RenameOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (h RenameOverlay) Margin() (int, int) {
	return 0, 0
}

func (h RenameOverlay) Title() string {
	return "Rename File"
}

func (h RenameOverlay) Init() tea.Cmd {
	return textinput.Blink
}

func (h RenameOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			cmds = append(cmds, overlay.Close(RenameOverlayID))
		case key.Matches(msg, config.Keys.OK, config.Keys.Editor.RenameFile):
			cmds = append(cmds, tea.Sequence(
				overlay.Close(RenameOverlayID),
				file.RenameFile(h.fileName.Value()),
			))
			return h, tea.Batch(cmds...)
		}
	}

	var cmd tea.Cmd
	h.fileName, cmd = h.fileName.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return h, tea.Batch(cmds...)
}

func (h RenameOverlay) View(width int, height int) string {
	return h.fileName.View()
}
