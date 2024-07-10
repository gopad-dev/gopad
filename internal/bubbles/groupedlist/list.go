package groupedlist

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

func New(items []Item) Model {
	return Model{
		TextInput: textinput.New(),
		Keys:      DefaultKeyMap,
		Styles:    DefaultStyles,
		filter:    filterItems,
		items:     items,
	}
}

type Model struct {
	TextInput textinput.Model
	Keys      KeyMap
	Styles    Styles

	filter func(filter string, item Item) bool
	width  int
	height int
	items  []Item
	item   int
	offset int
}

func (m *Model) Items() []Item {
	return m.items
}

func (m *Model) SetItems(items []Item) {
	m.items = items
}

func (m Model) SelectedItem() Item {
	var i int
	var walk func(items []Item) Item
	walk = func(items []Item) Item {
		for _, item := range items {
			if i == m.item {
				return item
			}
			i++
			if subItems := item.Items(); len(subItems) > 0 {
				if found := walk(subItems); found != nil {
					return found
				}
			}
		}
		return nil
	}

	return walk(m.items)
}

func (m Model) Update(ctx tea.Context, msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

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
	m.TextInput, cmd = m.TextInput.Update(ctx, msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View(ctx tea.Context) string {
	var listWidth int
	if m.width > 0 {
		m.TextInput.Width = m.width - 2
		listWidth = m.width
	}

	listHeight := len(m.items)
	if m.height > 0 {
		listHeight = m.height - 1 // for the text input
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.TextInput.View(ctx),
		m.itemsView(ctx, listWidth, listHeight),
	)
}

func (m *Model) itemsView(ctx tea.Context, width int, height int) string {
	var list string

	var i int
	var walk func(items []Item, indent int)
	walk = func(items []Item, indent int) {
		for _, item := range items {
			list += m.itemView(i, item, 0)
			i++
			if subItems := item.Items(); len(subItems) > 0 {
				walk(subItems, indent+1)
			}
		}
	}

	walk(m.items, 0)
	return list
}

func (m *Model) itemView(i int, item Item, indent int) string {
	style := m.Styles.ItemStyle
	if i == m.item {
		style = m.Styles.ItemSelectedStyle
	}

	return style.PaddingLeft(indent).Render(item.Title())
}
