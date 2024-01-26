package editor

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const DeleteOverlayID = "editor.delete"

var _ overlay.Overlay = (*DeleteOverlay)(nil)

func NewDeleteOverlay() DeleteOverlay {
	bOK := config.NewButton("OK", func() tea.Cmd {
		return tea.Sequence(overlay.Close(DeleteOverlayID), Delete)
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

func (q DeleteOverlay) ID() string {
	return DeleteOverlayID
}

func (q DeleteOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (q DeleteOverlay) Margin() (int, int) {
	return 0, 0
}

func (q DeleteOverlay) Title() string {
	return "Delete File"
}

func (q DeleteOverlay) Init() tea.Cmd {
	return nil
}

func (q DeleteOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return q, overlay.Close(DeleteOverlayID)
		case key.Matches(msg, config.Keys.Editor.DeleteRight):
			return q, tea.Sequence(overlay.Close(DeleteOverlayID), Delete)
		case key.Matches(msg, config.Keys.Left):
			q.buttonOK.Focus()
			q.buttonCancel.Blur()
		case key.Matches(msg, config.Keys.Right):
			q.buttonOK.Blur()
			q.buttonCancel.Focus()
		}
	}

	var cmd tea.Cmd
	q.buttonOK, cmd = q.buttonOK.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	q.buttonCancel, cmd = q.buttonCancel.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return q, tea.Batch(cmds...)
}

func (q DeleteOverlay) View(width int, height int) string {
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginBottom(1).Render("Are you sure you want to delete this file?"),
		lipgloss.JoinHorizontal(lipgloss.Center, q.buttonOK.View(), q.buttonCancel.View()),
	)
}
