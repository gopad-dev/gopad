package editor

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	file2 "go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const CloseOverlayID = "editor.close"

var _ overlay.Overlay = (*CloseOverlay)(nil)

func NewCloseOverlay(files []string) CloseOverlay {
	bOK := config.NewButton("OK", func() tea.Cmd {
		var cmds []tea.Cmd
		for _, file := range files {
			cmds = append(cmds, file2.CloseFile(file))
		}
		cmds = append(cmds, overlay.Close(CloseOverlayID))
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

func (o CloseOverlay) ID() string {
	return CloseOverlayID
}

func (o CloseOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (o CloseOverlay) Margin() (int, int) {
	return 0, 0
}

func (o CloseOverlay) Title() string {
	if len(o.files) > 1 {
		return "Close Files"
	}
	return "Close File"
}

func (o CloseOverlay) Init() tea.Cmd {
	return nil
}

func (o CloseOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.CloseFile):
			for _, file := range o.files {
				cmds = append(cmds, file2.CloseFile(file))
			}
			return o, tea.Sequence(cmds...)
		case key.Matches(msg, config.Keys.Left):
			o.buttonOK.Focus()
			o.buttonCancel.Blur()
		case key.Matches(msg, config.Keys.Right):
			o.buttonOK.Blur()
			o.buttonCancel.Focus()
		case key.Matches(msg, config.Keys.Cancel):
			return o, overlay.Close(CloseOverlayID)
		}
	}

	var cmd tea.Cmd
	o.buttonOK, cmd = o.buttonOK.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	o.buttonCancel, cmd = o.buttonCancel.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return o, tea.Batch(cmds...)
}

func (o CloseOverlay) View(width int, height int) string {
	msg := "You have unsaved changes. Are you sure you want to close?"
	if len(o.files) > 1 {
		msg = fmt.Sprintf("You have unsaved changes in %d files. Are you sure you want to close?", len(o.files))
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginBottom(1).Render(msg),
		lipgloss.JoinHorizontal(lipgloss.Center, o.buttonOK.View(), o.buttonCancel.View()),
	)
}
