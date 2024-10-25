package editor

import (
	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/key"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const DeleteOverlayID = "editor.delete"

var _ overlay.Overlay = (*DeleteOverlay)(nil)

func NewDeleteOverlay(files []string) DeleteOverlay {
	bOK := config.NewButton("OK", func() tea.Cmd {
		cmds := []tea.Cmd{
			overlay.Close(DeleteOverlayID),
		}
		for _, f := range files {
			cmds = append(cmds, file.DeleteFile(f))
		}
		return tea.Sequence(cmds...)
	})

	bCancel := config.NewButton("Cancel", func() tea.Cmd {
		return overlay.Close(DeleteOverlayID)
	})
	bCancel.Focus()

	return DeleteOverlay{
		files:        files,
		buttonOK:     bOK,
		buttonCancel: bCancel,
	}
}

type DeleteOverlay struct {
	files []string

	buttonOK     button.Model
	buttonCancel button.Model
}

func (d DeleteOverlay) ID() string {
	return DeleteOverlayID
}

func (d DeleteOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (d DeleteOverlay) Margin() (int, int) {
	return 0, 0
}

func (d DeleteOverlay) Title() string {
	if len(d.files) > 1 {
		return "Delete Files"
	}
	return "Delete File"
}

func (d DeleteOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return d, nil
}

func (d DeleteOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.File.Delete):
			return d, d.buttonOK.OnClick()
		case key.Matches(msg, config.Keys.Cancel):
			return d, d.buttonCancel.OnClick()
		case key.Matches(msg, config.Keys.Left):
			d.buttonOK.Focus()
			d.buttonCancel.Blur()
			return d, nil
		case key.Matches(msg, config.Keys.Right):
			d.buttonOK.Blur()
			d.buttonCancel.Focus()
			return d, nil
		}
	}

	var cmd tea.Cmd
	d.buttonOK, cmd = d.buttonOK.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	d.buttonCancel, cmd = d.buttonCancel.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return d, tea.Batch(cmds...)
}

func (d DeleteOverlay) View(width int, height int) string {
	msg := "Are you sure you want to delete this file?"
	if len(d.files) > 1 {
		msg = "Are you sure you want to delete these files?"
	}
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginBottom(1).Render(msg),
		lipgloss.JoinHorizontal(lipgloss.Center, d.buttonOK.View(), d.buttonCancel.View()),
	)
}
