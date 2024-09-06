package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

const GoToOverlayID = "editor.goto"

var _ overlay.Overlay = (*GoToOverlay)(nil)

func NewGoToOverlay(row int, col int) GoToOverlay {
	ti := textinput.New()
	ti.Placeholder = "[line] [:column]"
	ti.SetValue(fmt.Sprintf("%d:%d", row, col))
	return GoToOverlay{
		ti: ti,
	}
}

type GoToOverlay struct {
	ti textinput.Model
}

func (o GoToOverlay) ID() string {
	return GoToOverlayID
}

func (o GoToOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (o GoToOverlay) Margin() (int, int) {
	return 0, 0
}

func (o GoToOverlay) Title() string {
	return "Go To"
}

func (o GoToOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return o, o.ti.Focus()
}

func (o GoToOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return o, overlay.Close(GoToOverlayID)
		case key.Matches(msg, config.Keys.OK):
			pos := strings.SplitN(o.ti.Value(), ":", 2)
			row, err := strconv.Atoi(pos[0])
			if err != nil {
				cmds = append(cmds, notifications.Add(fmt.Sprintf("Error parsing row: %s", err)))
				return o, tea.Batch(cmds...)
			}
			col, err := strconv.Atoi(pos[1])
			if err != nil {
				cmds = append(cmds, notifications.Add(fmt.Sprintf("Error parsing column: %s", err)))
				return o, tea.Batch(cmds...)
			}

			cmds = append(cmds, tea.Sequence(
				overlay.Close(GoToOverlayID),
				file.Scroll(row, col),
			))
			return o, tea.Batch(cmds...)
		}
	}

	var cmd tea.Cmd
	o.ti, cmd = o.ti.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return o, tea.Batch(cmds...)
}

func (o GoToOverlay) View(_ int, _ int) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		o.ti.View(),
		"Press [esc] to cancel or [enter] to go to.",
	)
}
