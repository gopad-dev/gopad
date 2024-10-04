package list

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/lrstanley/bubblezone"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

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

func New[T Item](items []T) Model[T] {
	ti := textinput.New()
	ti.Placeholder = "Type to search"

	return Model[T]{
		TextInput:  ti,
		KeyMap:     DefaultKeyMap,
		Styles:     DefaultStyles,
		items:      parseItems(items),
		zonePrefix: zone.NewPrefix(),
	}
}

type Model[T Item] struct {
	TextInput  textinput.Model
	KeyMap     KeyMap
	Styles     Styles
	width      int
	height     int
	items      []modelItem[T]
	item       int
	offset     int
	zonePrefix string
	clicked    bool
}

func (m *Model[T]) Focus() tea.Cmd {
	return m.TextInput.Focus()
}

func (m *Model[T]) Blur() {
	m.TextInput.Blur()
}

func (m *Model[T]) Focused() bool {
	return m.TextInput.Focused()
}

func (m *Model[T]) SetWidth(width int) {
	m.width = width
}

func (m *Model[T]) SetHeight(height int) {
	m.height = height
}

func (m *Model[T]) selectedItem() *modelItem[T] {
	items := m.filteredItems()
	if m.item < len(items) {
		return &items[m.item]
	}
	return nil
}

func (m *Model[T]) Selected() T {
	if item := m.selectedItem(); item != nil {
		return item.item
	}
	var zero T
	return zero
}

func (m *Model[T]) Select(i int) {
	if i >= 0 && i < len(m.items) {
		m.item = i
	}
}

func (m *Model[T]) SelectedIndex() int {
	selected := m.selectedItem()
	if selected != nil {
		return selected.index
	}
	return -1
}

func (m *Model[T]) Items() []T {
	return modelItems(m.filteredItems())
}

func (m *Model[T]) AllItems() []T {
	return modelItems(m.items)
}

func (m *Model[T]) SetItems(items []T) {
	m.items = parseItems(items)
}

func (m *Model[T]) filteredItems() []modelItem[T] {
	return filterItems(m.items, m.TextInput.Value())
}

func (m Model[T]) zoneItemID(i int) string {
	return fmt.Sprintf("list:%s:%d", m.zonePrefix, i)
}

func (m Model[T]) zoneID() string {
	return fmt.Sprintf("list:%s", m.zonePrefix)
}

func (m *Model[T]) Clicked() bool {
	return m.clicked
}

func (m Model[T]) Update(msg tea.Msg) (Model[T], tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		for i := range len(m.filteredItems()) {
			if mouse.Matches(msg, m.zoneItemID(i), tea.MouseLeft /*, tea.MouseActionRelease*/) {
				m.item = i
				m.clicked = true
				return m, nil
			}
		}

		switch {
		case mouse.Matches(msg, m.zoneID(), tea.MouseWheelUp):
			if m.item > 0 {
				m.item--
			}
			return m, nil
		case mouse.Matches(msg, m.zoneID(), tea.MouseWheelDown):
			if m.item < len(m.items)-1 {
				m.item++
			}
			return m, nil
		}
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			if m.item > 0 {
				m.item--
			}
			return m, nil
		case key.Matches(msg, m.KeyMap.Down):
			if m.item < len(m.items)-1 {
				m.item++
			}
			return m, nil
		case key.Matches(msg, m.KeyMap.Start):
			m.item = 0
			return m, nil
		case key.Matches(msg, m.KeyMap.End):
			m.item = len(m.items) - 1
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)

	return m, cmd
}

// calculateOffset calculates the offset based on the current selected modelItem and the height of the list
func (m *Model[T]) calculateOffset() {
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

func (m Model[T]) View() string {
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

func (m Model[T]) itemsView(width int, height int) string {
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

		strs := []string{item.item.Title()}
		if item.item.Description() != "" {
			strs = append(strs, m.Styles.ItemDescriptionStyle.Render(item.item.Description()))
		}

		s := style.Render(strs...)
		if m.zonePrefix != "" {
			s = zone.Mark(m.zoneItemID(ii), s)
		}

		str += s + "\n"
	}

	str = strings.TrimRight(str, "\n")

	style := m.Styles.Style
	if width > 0 {
		style = style.Width(width - style.GetHorizontalFrameSize())
	}

	str = style.Render(str)

	if m.zonePrefix != "" {
		str = zone.Mark(m.zoneID(), str)
	}

	return str
}
