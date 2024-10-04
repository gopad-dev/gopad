package button

import (
	"fmt"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/lrstanley/bubblezone"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/internal/bubbles/mouse"
)

type KeyMap struct {
	OK key.Binding
}

var DefaultKeyMap = KeyMap{
	OK: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "ok"),
	),
}

type Styles struct {
	Default lipgloss.Style
	Focus   lipgloss.Style
}

var DefaultStyles = Styles{
	Default: lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Reverse(true),
	Focus:   lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Foreground(lipgloss.ANSIColor(13)).Reverse(true),
}

func New(label string, onclick func() tea.Cmd) Model {
	return Model{
		Label:   label,
		OnClick: onclick,

		Styles:     DefaultStyles,
		KeyMap:     DefaultKeyMap,
		zonePrefix: zone.NewPrefix(),
	}
}

type Model struct {
	Label   string
	OnClick func() tea.Cmd

	Styles Styles
	KeyMap KeyMap

	focus      bool
	zonePrefix string
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) Blur() {
	m.focus = false
}

func (m Model) zoneID() string {
	return fmt.Sprintf("button:%s", m.zonePrefix)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.MouseReleaseMsg:
		switch {
		case mouse.Matches((msg), m.zoneID(), tea.MouseLeft):
			cmds = append(cmds, m.OnClick())
			return m, tea.Batch(cmds...)
		}
	case tea.KeyPressMsg:
		if m.focus {
			switch {
			case key.Matches(msg, m.KeyMap.OK):
				cmds = append(cmds, m.OnClick())
				return m, tea.Batch(cmds...)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.focus {
		return zone.Mark(m.zoneID(), m.Styles.Focus.Render(m.Label))
	}

	return zone.Mark(m.zoneID(), m.Styles.Default.Render(m.Label))
}
