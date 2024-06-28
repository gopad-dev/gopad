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

const NewOverlayID = "editor.new"

var _ overlay.Overlay = (*NewOverlay)(nil)

func NewNewOverlay() NewOverlay {
	ti := config.NewTextInput()
	ti.Placeholder = "New file name"
	ti.Focus()

	return NewOverlay{
		fileName: ti,
	}
}

type NewOverlay struct {
	fileName textinput.Model
}

func (o NewOverlay) ID() string {
	return NewOverlayID
}

func (o NewOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (o NewOverlay) Margin() (int, int) {
	return 0, 0
}

func (o NewOverlay) Title() string {
	return "New File"
}

func (o NewOverlay) Init() tea.Cmd {
	return textinput.Blink
}

func (o NewOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel, config.Keys.Editor.NewFile):
			cmds = append(cmds, overlay.Close(NewOverlayID))
		case key.Matches(msg, config.Keys.OK):
			cmds = append(cmds, tea.Sequence(
				overlay.Close(NewOverlayID),
				file.NewFile(o.fileName.Value()),
			))
			return o, tea.Batch(cmds...)
		}
	}

	var cmd tea.Cmd
	o.fileName, cmd = o.fileName.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return o, tea.Batch(cmds...)
}

func (o NewOverlay) View(width int, height int) string {
	style := config.Theme.Overlay.Styles.Style
	width /= 2
	width -= style.GetHorizontalFrameSize()
	if width > 0 {
		o.fileName.Width = width - 4
	}

	return o.fileName.View()
}
