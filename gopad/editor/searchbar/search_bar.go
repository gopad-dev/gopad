package searchbar

import (
	"fmt"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/lrstanley/bubblezone"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/editormsg"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

const ZoneID = "editor.search-bar"

func onSelect(result Result) tea.Cmd {
	return file.Scroll(result.RowStart, result.ColStart)
}

type Result struct {
	RowStart int
	ColStart int
	RowEnd   int
	ColEnd   int
}

func New() Model {
	ti := config.NewTextInput()
	ti.Placeholder = "type to search"
	ti.Width = 20

	return Model{
		TextInput: ti,
	}
}

type Model struct {
	TextInput textinput.Model
	focus     bool
	show      bool

	results     []Result
	resultIndex int
}

func (m *Model) Visible() bool {
	return m.show
}

func (m *Model) Show() {
	m.show = true
}

func (m *Model) Hide() {
	m.show = false
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus() tea.Cmd {
	m.focus = true
	return m.TextInput.Focus()
}

func (m *Model) Blur() {
	m.focus = false
	m.TextInput.Blur()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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
				cmds = append(cmds, editormsg.Focus(editormsg.ModelSearch))
			}
			return m, tea.Batch(cmds...)
		}
	case tea.KeyPressMsg:
		if m.Focused() {
			switch {
			case key.Matches(msg, config.Keys.Editor.SearchBar.Close):
				m.Hide()
				cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
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
					cmds = append(cmds, onSelect(m.results[m.resultIndex]), editormsg.Focus(editormsg.ModelFile))
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

func (m Model) View() string {
	results := "0 results"
	if len(m.results) > 0 {
		results = fmt.Sprintf(" %d/%d 🠅🠇", m.resultIndex+1, len(m.results))
	}

	return config.Theme.UI.SearchBar.Style.Render(zone.Mark(ZoneID, lipgloss.JoinHorizontal(lipgloss.Center,
		m.TextInput.View(),
		config.Theme.UI.SearchBar.ResultStyle.Render(results),
	)))
}
