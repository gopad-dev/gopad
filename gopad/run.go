package gopad

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

const RunOverlayID = "run"

var fallbackShells = map[string]string{
	"windows": "cmd",
	"darwin":  "sh",
	"linux":   "sh",
}

func Terminal() tea.Cmd {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = fallbackShells[runtime.GOOS]
	}
	return tea.ExecProcess(exec.Command(shell), func(err error) tea.Msg {
		if err != nil {
			return notifications.Add(fmt.Sprintf("error while executing command: %s", err.Error()))
		}
		return nil
	})
}

func NewRunOverlay() RunOverlay {
	l := config.NewList(Actions)
	l.TextInput.Placeholder = "Type a command and press enter to run it"
	l.Focus()

	return RunOverlay{
		list: l,
	}
}

type RunOverlay struct {
	list list.Model[Action]
}

func (r RunOverlay) ID() string {
	return RunOverlayID
}

func (r RunOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Top
}

func (r RunOverlay) Margin() (int, int) {
	return 0, 2
}

func (r RunOverlay) Title() string {
	return "Run"
}

func (r RunOverlay) Init() tea.Cmd {
	return textinput.Blink
}

func (r RunOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.OK):
			item := r.list.Selected()
			if item == nil {
				return r, nil
			}

			return r, tea.Batch(overlay.Close(RunOverlayID), item.(Action).Run())
		case key.Matches(msg, config.Keys.Cancel):
			return r, overlay.Close(RunOverlayID)
		}
	}

	var cmd tea.Cmd
	r.list, cmd = r.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return r, tea.Batch(cmds...)
}

func (r RunOverlay) View(width int, height int) string {
	style := config.Theme.Overlay.RunOverlayStyle
	width /= 2
	width -= style.GetHorizontalFrameSize()
	if width > 0 {
		r.list.SetWidth(width)
	}

	r.list.SetHeight(height - style.GetVerticalFrameSize() - 2)
	return r.list.View()
}
