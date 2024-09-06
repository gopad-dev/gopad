package editor

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const CloseOverlayID = "editor.close"

var _ overlay.Overlay = (*CloseOverlay)(nil)

func NewCloseOverlay(files []string) CloseOverlay {
	bOK := config.NewButton("OK", func() tea.Cmd {
		cmds := []tea.Cmd{
			overlay.Close(CloseOverlayID),
		}
		for _, f := range files {
			cmds = append(cmds, file.CloseFile(f))
		}
		return tea.Sequence(cmds...)
	})

	bCancel := config.NewButton("Cancel", func() tea.Cmd {
		return overlay.Close(CloseOverlayID)
	})
	bCancel.Focus()

	return CloseOverlay{
		files:        files,
		buttonOK:     bOK,
		buttonCancel: bCancel,
	}
}

type CloseOverlay struct {
	files []string

	buttonOK     button.Model
	buttonCancel button.Model
}

func (c CloseOverlay) ID() string {
	return CloseOverlayID
}

func (c CloseOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (c CloseOverlay) Margin() (int, int) {
	return 0, 0
}

func (c CloseOverlay) Title() string {
	if len(c.files) > 1 {
		return "Close Files"
	}
	return "Close File"
}

func (c CloseOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return c, nil
}

func (c CloseOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.File.Close):
			return c, c.buttonOK.OnClick()
		case key.Matches(msg, config.Keys.Cancel):
			return c, overlay.Close(CloseOverlayID)
		case key.Matches(msg, config.Keys.Left):
			c.buttonOK.Focus()
			c.buttonCancel.Blur()
			return c, nil
		case key.Matches(msg, config.Keys.Right):
			c.buttonOK.Blur()
			c.buttonCancel.Focus()
			return c, nil
		}
	}

	var cmd tea.Cmd
	c.buttonOK, cmd = c.buttonOK.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	c.buttonCancel, cmd = c.buttonCancel.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}

func (c CloseOverlay) View(width int, height int) string {
	msg := "You have unsaved changes. Are you sure you want to close?"
	if len(c.files) > 1 {
		msg = fmt.Sprintf("You have unsaved changes in %d files. Are you sure you want to close?", len(c.files))
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginBottom(1).Render(msg),
		lipgloss.JoinHorizontal(lipgloss.Center, c.buttonOK.View(), c.buttonCancel.View()),
	)
}
