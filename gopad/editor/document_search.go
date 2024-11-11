package editor

import (
	"fmt"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/lrstanley/bubblezone"

	"go.gopad.dev/gopad/gopad/editor/buffer"

	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

const ZoneID = "editor.search-bar"

func onSelect(result Result) tea.Cmd {
	return ScrollAction(result.Start)
}

type Result buffer.Range

func newSearchBar() searchBar {
	ti := config.NewTextInput()
	ti.Placeholder = "type to search"
	ti.Width = 20

	return searchBar{
		TextInput: ti,
	}
}

type searchBar struct {
	TextInput textinput.Model
	focus     bool
	show      bool

	results     []Result
	resultIndex int
}

func (m *searchBar) Visible() bool {
	return m.show
}

func (m *searchBar) Show() {
	m.show = true
}

func (m *searchBar) Hide() {
	m.show = false
}

func (m *searchBar) Focused() bool {
	return m.focus
}

func (m *searchBar) Focus() tea.Cmd {
	m.focus = true
	return m.TextInput.Focus()
}

func (m *searchBar) Blur() {
	m.focus = false
	m.TextInput.Blur()
}

func (m searchBar) Update(msg tea.Msg) (searchBar, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case searchResultMsg:
		m.results = msg.Results
		m.resultIndex = 0
		if len(m.results) > 0 {
			cmds = append(cmds, onSelect(m.results[0]))
		}
	case tea.MouseMsg:
		switch {
		case mouse.Matches(msg, ZoneID, tea.MouseLeft):
			if !m.Focused() {
				cmds = append(cmds, Focus(ModelTypeSearchBar))
			}
			return m, tea.Batch(cmds...)
		}
	case tea.KeyPressMsg:
		if m.Focused() {
			switch {
			case key.Matches(msg, config.Keys.Editor.SearchBar.Close):
				m.Hide()
				cmds = append(cmds, Focus(ModelTypeFile))
				return m, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.SearchBar.SelectPrev):
				if len(m.results) > 0 {
					if m.resultIndex > 0 {
						m.resultIndex--
					} else {
						m.resultIndex = len(m.results) - 1
					}
					cmds = append(cmds, onSelect(m.results[m.resultIndex]))
				}
			case key.Matches(msg, config.Keys.Editor.SearchBar.SelectNext):
				if len(m.results) > 0 {
					if m.resultIndex < len(m.results)-1 {
						m.resultIndex++
					} else {
						m.resultIndex = 0
					}
					cmds = append(cmds, onSelect(m.results[m.resultIndex]))
				}
			case key.Matches(msg, config.Keys.Editor.SearchBar.SelectResult):
				if len(m.results) > 0 {
					cmds = append(cmds, onSelect(m.results[m.resultIndex]), Focus(ModelTypeFile))
				}
				return m, tea.Batch(cmds...)
			}
		}
	}

	if m.Focused() {
		previousValue := m.TextInput.Value()
		var cmd tea.Cmd
		m.TextInput, cmd = m.TextInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		newValue := m.TextInput.Value()
		if previousValue != newValue && newValue != "" {
			cmds = append(cmds, Search(newValue))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m searchBar) View() string {
	results := "0 results"
	if len(m.results) > 0 {
		results = fmt.Sprintf(" %d/%d 🠅🠇", m.resultIndex+1, len(m.results))
	}

	return config.Theme.UI.SearchBar.Style.Render(zone.Mark(ZoneID, lipgloss.JoinHorizontal(lipgloss.Center,
		m.TextInput.View(),
		config.Theme.UI.SearchBar.ResultStyle.Render(results),
	)))
}
