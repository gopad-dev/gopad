package editor

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const DeleteOverlayID = "editor.delete"

var _ overlay.Overlay = (*DeleteOverlay)(nil)

func NewDeleteOverlay() DeleteOverlay {
	bOK := config.NewButton("OK", func() tea.Cmd {
		return tea.Sequence(overlay.Close(DeleteOverlayID), file.Delete)
	})

	bCancel := config.NewButton("Cancel", func() tea.Cmd {
		return overlay.Close(DeleteOverlayID)
	})
	bCancel.Focus()

	return DeleteOverlay{
		buttonOK:     bOK,
		buttonCancel: bCancel,
	}
}

type DeleteOverlay struct {
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
	return "Delete File"
}

func (d DeleteOverlay) Init() tea.Cmd {
	return nil
}

func (d DeleteOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.DeleteFile):
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
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginBottom(1).Render("Are you sure you want to delete this file?"),
		lipgloss.JoinHorizontal(lipgloss.Center, d.buttonOK.View(), d.buttonCancel.View()),
	)
}
