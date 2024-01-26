package list

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

type Item interface {
	Title() string
	Description() string
}

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Start key.Binding
	End   key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "select previous entry"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "select next entry"),
	),
	Start: key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "select first entry"),
	),
	End: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "select last entry"),
	),
}

var DefaultStyles = Styles{
	Style:                lipgloss.NewStyle().MarginLeft(1),
	ItemStyle:            lipgloss.NewStyle().Padding(0, 1),
	ItemSelectedStyle:    lipgloss.NewStyle().Padding(0, 1).Reverse(true),
	ItemDescriptionStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
}

type Styles struct {
	Style                lipgloss.Style
	ItemStyle            lipgloss.Style
	ItemSelectedStyle    lipgloss.Style
	ItemDescriptionStyle lipgloss.Style
}

func New(items []Item) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to search"

	return Model{
		TextInput: ti,
		Keys:      DefaultKeyMap,
		Styles:    DefaultStyles,
		items:     items,
	}
}

type Model struct {
	TextInput textinput.Model
	Keys      KeyMap
	Styles    Styles
	width     int
	height    int
	items     []Item
	item      int
	offset    int
}

func (m *Model) Focus() tea.Cmd {
	return m.TextInput.Focus()
}

func (m *Model) Blur() {
	m.TextInput.Blur()
}

func (m *Model) Focused() bool {
	return m.TextInput.Focused()
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) SetHeight(height int) {
	m.height = height
}

func (m *Model) Selected() Item {
	items := m.filteredItems()
	if m.item < len(items) {
		return items[m.item]
	}
	return nil
}

func (m *Model) Select(i int) {
	if i >= 0 && i < len(m.items) {
		m.item = i
	}
}

func (m *Model) SelectedIndex() int {
	return m.item
}

func (m *Model) Items() []Item {
	return m.items
}

func (m *Model) SetItems(items []Item) {
	m.items = items
}

func (m *Model) filteredItems() []Item {
	value := strings.ToLower(m.TextInput.Value())
	if value == "" {
		return m.items
	}

	var filtered []Item
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item.Title()), value) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Up):
			if m.item > 0 {
				m.item--
			}
			return m, nil
		case key.Matches(msg, m.Keys.Down):
			if m.item < len(m.items)-1 {
				m.item++
			}
			return m, nil
		case key.Matches(msg, m.Keys.Start):
			m.item = 0
			return m, nil
		case key.Matches(msg, m.Keys.End):
			m.item = len(m.items) - 1
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)

	return m, cmd
}

// calculateOffset calculates the offset based on the current selected item and the height of the list
func (m *Model) calculateOffset() {
	if m.height == 0 {
		m.offset = 0
		return
	}

	if m.item >= m.offset+m.height {
		m.offset = m.item - m.height + 1
	} else if m.item < m.offset {
		m.offset = m.item
	}
}

func (m Model) View() string {
	m.calculateOffset()
	var listWidth int
	if m.width > 0 {
		m.TextInput.Width = m.width - 2
		listWidth = m.width
	}

	listHeight := len(m.items)
	if m.height > 0 {
		listHeight = m.height // -1 for the text input
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.TextInput.View(),
		m.itemsView(listWidth, listHeight),
	)
}

func (m Model) itemsView(width int, height int) string {
	var str string

	items := m.filteredItems()
	if m.item >= len(items) {
		m.item = len(items) - 1
	}

	for i := range height {
		ii := i + m.offset
		if ii >= len(items) {
			break
		}

		item := items[ii]
		style := m.Styles.ItemStyle
		if ii == m.item {
			style = m.Styles.ItemSelectedStyle
		}

		strs := []string{item.Title()}
		if item.Description() != "" {
			strs = append(strs, m.Styles.ItemDescriptionStyle.Render(item.Description()))
		}

		str += style.Render(strs...) + "\n"
	}

	str = strings.TrimRight(str, "\n")

	style := m.Styles.Style
	if width > 0 {
		style = style.Width(width - style.GetHorizontalFrameSize())
	}

	return style.Render(str)
}
