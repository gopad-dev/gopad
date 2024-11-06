package gopad

import (
	"log"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/key"
	"go.gopad.dev/gopad/internal/bubbles/label"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/textinput"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const LSPOverlayID = "lsp"

var _ overlay.Overlay = (*LSPOverlay)(nil)

var style = lipgloss.NewStyle().PaddingTop(1)

func NewLSPOverlay(servers []ls.ServerConfig, workspace string) LSPOverlay {
	l := config.NewList(servers)
	l.TextInput.Placeholder = "Type a language server and press enter to enable it"
	l.Focus()

	t := config.NewTextInput()
	t.Placeholder = "The workspace path"
	t.SetValue(workspace)

	ok := config.NewButton("OK", nil)
	cancel := config.NewButton("Cancel", nil)

	return LSPOverlay{
		listLabel:      config.NewLabel("Select a language server:"),
		list:           l,
		textinputLabel: config.NewLabel("Workspace path:"),
		textinput:      t,
		ok:             ok,
		cancel:         cancel,
		workspace:      workspace,
	}
}

type LSPOverlay struct {
	listLabel label.Model
	list      list.Model[ls.ServerConfig]

	textinputLabel label.Model
	textinput      textinput.Model

	ok     button.Model
	cancel button.Model

	workspace string
	focus     int
}

func (o LSPOverlay) ID() string {
	return LSPOverlayID
}

func (o LSPOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Center
}

func (o LSPOverlay) Margin() (int, int) {
	return 0, 0
}

func (o LSPOverlay) Title() string {
	return "LSP Overlay"
}

func (o LSPOverlay) Init() (overlay.Overlay, tea.Cmd) {
	return o, nil
}

func (o *LSPOverlay) FocusNext() tea.Cmd {
	o.focus++
	if o.focus > 3 {
		o.focus = 0
	}
	return o.SetFocus(o.focus)
}

func (o *LSPOverlay) FocusPrev() tea.Cmd {
	o.focus--
	if o.focus < 0 {
		o.focus = 3
	}
	return o.SetFocus(o.focus)
}

func (o *LSPOverlay) SetFocus(f int) tea.Cmd {
	o.focus = f

	var cmds []tea.Cmd
	switch f {
	case 0:
		cmds = append(cmds, o.list.Focus())
		o.textinput.Blur()
		o.ok.Blur()
		o.cancel.Blur()
	case 1:
		o.list.Blur()
		cmds = append(cmds, o.textinput.Focus())
		o.ok.Blur()
		o.cancel.Blur()
	case 2:
		o.list.Blur()
		o.textinput.Blur()
		o.ok.Focus()
		o.cancel.Blur()
	case 3:
		o.list.Blur()
		o.textinput.Blur()
		o.ok.Blur()
		o.cancel.Focus()
	}

	return tea.Batch(cmds...)
}

func (o LSPOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.FocusNext):
			log.Println("FocusNext")
			cmds = append(cmds, o.FocusNext())
			return o, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.FocusPrev):
			log.Println("FocusPrev")
			cmds = append(cmds, o.FocusPrev())
			return o, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Cancel):
			cmds = append(cmds,
				overlay.Close(LSPOverlayID),
				ls.StartServer(o.list.Selected().Name, o.workspace),
			)
			return o, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.OK):
			item := o.list.Selected()
			if item.Name != "" {
				cmds = append(cmds,
					overlay.Close(LSPOverlayID),
					ls.StartServer(o.list.Selected().Name, o.workspace),
				)
				return o, tea.Batch(cmds...)
			}
		}
	}

	var cmd tea.Cmd
	switch o.focus {
	case 0:
		o.list, cmd = o.list.Update(msg)
	case 1:
		o.textinput, cmd = o.textinput.Update(msg)
	case 2:
		o.ok, cmd = o.ok.Update(msg)
	case 3:
		o.cancel, cmd = o.cancel.Update(msg)
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return o, tea.Batch(cmds...)
}

func (o LSPOverlay) View(width int, height int) string {
	o.list.SetHeight(height)
	o.list.SetWidth(width / 2)
	o.textinput.Width = width / 2

	return lipgloss.JoinVertical(lipgloss.Left,
		o.listLabel.View(o.list.View()),
		style.Render(o.textinputLabel.View(o.textinput.View())),
		style.
			Width(width/2).
			AlignVertical(lipgloss.Center).
			Render(
				o.ok.View(),
				o.cancel.View(),
			),
	)
}
