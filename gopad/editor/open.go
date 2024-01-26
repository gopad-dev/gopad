package editor

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const OpenOverlayID = "editor.open"

var _ overlay.Overlay = (*OpenOverlay)(nil)

func NewOpenOverlay(dir string, fileAllowed bool, dirAllowed bool) OpenOverlay {
	fp := config.NewFilePicker(dir, fileAllowed, dirAllowed)
	return OpenOverlay{
		filePicker: fp,
	}
}

type OpenOverlay struct {
	filePicker filepicker.Model
}

func (o OpenOverlay) ID() string {
	return OpenOverlayID
}

func (o OpenOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (o OpenOverlay) Margin() (int, int) {
	return 0, 0
}

func (o OpenOverlay) Title() string {
	s := "Open"
	if o.filePicker.FileAllowed {
		s += " File"
	}
	if o.filePicker.DirAllowed {
		s += " Directory"
	}
	return s
}

func (o OpenOverlay) Init() tea.Cmd {
	return o.filePicker.Init()
}

func (o OpenOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return o, overlay.Close(OpenOverlayID)
		}
	}

	var cmd tea.Cmd
	o.filePicker, cmd = o.filePicker.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if ok, file := o.filePicker.DidSelect(msg); ok {
		stat, err := os.Stat(file)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("Error statting path %s: %s", file, err)))
			return o, tea.Batch(cmds...)
		}
		if stat.IsDir() {
			cmd = OpenDir(file)
		} else {
			cmd = OpenFile(file)
		}
		cmds = append(cmds, tea.Sequence(overlay.Close(OpenOverlayID), cmd))
	}
	return o, tea.Batch(cmds...)
}

func (o OpenOverlay) View(_ int, height int) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		o.filePicker.View(height-config.Theme.Overlay.Styles.Style.GetVerticalFrameSize()-4),
		"Press [esc] to cancel or [enter] to open.",
	)
}
