package gopad

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const QuitOverlayID = "quit"

var _ overlay.Overlay = (*QuitOverlay)(nil)

func Quit() tea.Msg {
	return quitMsg{}
}

type quitMsg struct{}

func NewQuitOverlay() QuitOverlay {
	bOK := config.NewButton("OK", func() tea.Cmd {
		return tea.Quit
	})

	bCancel := config.NewButton("Cancel", func() tea.Cmd {
		return overlay.Close(QuitOverlayID)
	})
	bCancel.Focus()

	return QuitOverlay{
		buttonOK:     bOK,
		buttonCancel: bCancel,
	}
}

type QuitOverlay struct {
	buttonOK     button.Model
	buttonCancel button.Model
}

func (q QuitOverlay) ID() string {
	return QuitOverlayID
}

func (q QuitOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (q QuitOverlay) Margin() (int, int) {
	return 0, 0
}

func (q QuitOverlay) Title() string {
	return "Quit"
}

func (q QuitOverlay) Init(ctx tea.Context) (overlay.Overlay, tea.Cmd) {
	return q, nil
}

func (q QuitOverlay) Update(ctx tea.Context, msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Quit):
			return q, q.buttonOK.OnClick()
		case key.Matches(msg, config.Keys.Cancel):
			return q, q.buttonCancel.OnClick()
		case key.Matches(msg, config.Keys.Left):
			q.buttonOK.Focus()
			q.buttonCancel.Blur()
			return q, nil
		case key.Matches(msg, config.Keys.Right):
			q.buttonOK.Blur()
			q.buttonCancel.Focus()
			return q, nil
		}
	}

	var cmd tea.Cmd
	q.buttonOK, cmd = q.buttonOK.Update(ctx, msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	q.buttonCancel, cmd = q.buttonCancel.Update(ctx, msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return q, tea.Batch(cmds...)
}

func (q QuitOverlay) View(ctx tea.Context, width int, height int) string {
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().MarginBottom(1).Render("You have unsaved changes. Are you sure you want to quit?"),
		lipgloss.JoinHorizontal(lipgloss.Center, q.buttonOK.View(ctx), q.buttonCancel.View(ctx)),
	)
}
