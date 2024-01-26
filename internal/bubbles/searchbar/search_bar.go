package searchbar

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

type Result struct {
	RowStart int
	ColStart int
	RowEnd   int
	ColEnd   int
}

type Styles struct {
	Style       lipgloss.Style
	ResultStyle lipgloss.Style
}

var DefaultStyles = Styles{
	Style:       lipgloss.NewStyle().Padding(0, 2),
	ResultStyle: lipgloss.NewStyle().Padding(0, 1),
}

type KeyMap struct {
	SelectPrev key.Binding
	SelectNext key.Binding

	SelectResult key.Binding
	Close        key.Binding
}

func (m KeyMap) FullHelpView() []help.KeyMapCategory {
	return []help.KeyMapCategory{
		{
			Category: "Search",
			Keys: []key.Binding{
				m.SelectPrev,
				m.SelectNext,
				m.SelectResult,
				m.Close,
			},
		},
	}
}

var DefaultKeyMap = KeyMap{
	SelectPrev: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "select prev"),
	),
	SelectNext: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "select next"),
	),
	SelectResult: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select result"),
	),
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close search"),
	),
}

func New(onSelect func(result Result) tea.Cmd, onBlur tea.Cmd) Model {
	ti := textinput.New()
	ti.Placeholder = "type to search"
	ti.Width = 20

	return Model{
		Styles:    DefaultStyles,
		KeyMap:    DefaultKeyMap,
		TextInput: ti,
		onSelect:  onSelect,
		onBlur:    onBlur,
	}
}

type Model struct {
	Styles    Styles
	KeyMap    KeyMap
	TextInput textinput.Model
	focus     bool
	show      bool
	onSelect  func(result Result) tea.Cmd
	onBlur    tea.Cmd

	results     []Result
	resultIndex int
}

func (s *Model) Visible() bool {
	return s.show
}

func (s *Model) Show() {
	s.show = true
}

func (s *Model) Hide() {
	s.show = false
}

func (s *Model) Focused() bool {
	return s.focus
}

func (s *Model) Focus() tea.Cmd {
	s.focus = true
	return s.TextInput.Focus()
}

func (s *Model) Blur() {
	s.focus = false
	s.TextInput.Blur()
}

func (s Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case searchResultMsg:
		s.results = msg.Results
		s.resultIndex = 0
		if len(s.results) > 0 {
			cmds = append(cmds, s.onSelect(s.results[0]))
		}
	case tea.KeyMsg:
		if s.focus {
			switch {
			case key.Matches(msg, s.KeyMap.SelectPrev):
				if len(s.results) > 0 {
					if s.resultIndex > 0 {
						s.resultIndex--
					} else {
						s.resultIndex = len(s.results) - 1
					}
					cmds = append(cmds, s.onSelect(s.results[s.resultIndex]))
				}
			case key.Matches(msg, s.KeyMap.SelectNext):
				if len(s.results) > 0 {
					if s.resultIndex < len(s.results)-1 {
						s.resultIndex++
					} else {
						s.resultIndex = 0
					}
					cmds = append(cmds, s.onSelect(s.results[s.resultIndex]))
				}
			case key.Matches(msg, s.KeyMap.Close):
				s.Hide()
				s.Blur()
				cmds = append(cmds, s.onBlur)
			case key.Matches(msg, s.KeyMap.SelectResult):
				if len(s.results) > 0 {
					s.Blur()
					cmds = append(cmds, s.onSelect(s.results[s.resultIndex]), s.onBlur)
				}
			}
		}
	}

	if s.focus {
		previousValue := s.TextInput.Value()
		var cmd tea.Cmd
		s.TextInput, cmd = s.TextInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		newValue := s.TextInput.Value()
		if previousValue != newValue && newValue != "" {
			cmds = append(cmds, Search(newValue))
		}
	}

	return s, tea.Batch(cmds...)
}

func (s Model) View() string {
	results := "0 results"
	if len(s.results) > 0 {
		results = fmt.Sprintf(" %d/%d 🠅🠇", s.resultIndex+1, len(s.results))
	}

	return s.Styles.Style.Render(lipgloss.JoinHorizontal(lipgloss.Center,
		s.TextInput.View(),
		s.Styles.ResultStyle.Render(results),
	))
}
