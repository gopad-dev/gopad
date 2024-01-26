package button

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	Default: lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Copy().Reverse(true),
	Focus:   lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Copy().Foreground(lipgloss.ANSIColor(13)).Reverse(true),
}

func New(label string, onclick func() tea.Cmd) Model {
	return Model{
		Label:   label,
		OnClick: onclick,

		Styles: DefaultStyles,
		KeyMap: DefaultKeyMap,
	}
}

type Model struct {
	Label   string
	OnClick func() tea.Cmd

	Styles Styles
	KeyMap KeyMap

	focus bool
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

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.OK):
			cmds = append(cmds, m.OnClick())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.focus {
		return m.Styles.Focus.Render(m.Label)
	}

	return m.Styles.Default.Render(m.Label)
}
