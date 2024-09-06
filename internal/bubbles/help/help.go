package help

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap is a map of keybindings used to generate help. Since it's an
// interface it can be any type, though struct or a map[string][]key.Binding
// are likely candidates.
//
// Note that if a key is disabled (via key.Binding.SetEnabled) it will not be
// rendered in the help view, so in theory generated help should self-manage.
type KeyMap interface {
	// HelpView returns an extended group of help items, grouped by columns.
	// The help bubble will render the help in the order in which the help
	// items are returned here.
	HelpView() []KeyMapCategory
}

type KeyMapCategory struct {
	Category string
	Keys     []key.Binding
}

// Styles is a set of available style definitions for the Help bubble.
type Styles struct {
	Ellipsis  lipgloss.Style
	Group     lipgloss.Style
	Header    lipgloss.Style
	Key       lipgloss.Style
	Desc      lipgloss.Style
	Separator lipgloss.Style
}

var (
	keyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A4A4A"))
	sepStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C"))
)

var DefaultStyles = Styles{
	Header:    lipgloss.NewStyle().Reverse(true).AlignHorizontal(lipgloss.Center),
	Ellipsis:  sepStyle,
	Key:       keyStyle,
	Desc:      descStyle,
	Separator: sepStyle,
}

// Model contains the state of the help view.
type Model struct {
	Styles Styles
}

// New creates a new help view with some useful defaults.
func New() Model {
	return Model{
		Styles: DefaultStyles,
	}
}

// Update helps satisfy the Bubble Tea Model interface. It's a no-op.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders help columns from a slice of key binding slices. Each top level slice entry renders into a column.
func (m Model) View(k KeyMap, width int, height int) string {
	groups := k.HelpView()
	if len(groups) == 0 {
		return ""
	}

	var outs []string
	for _, group := range groups {
		if !shouldRenderColumn(group) {
			continue
		}

		var (
			keys         []string
			descriptions []string
		)

		// Separate keys and descriptions into different slices
		for _, kb := range group.Keys {
			if !kb.Enabled() {
				continue
			}
			keys = append(keys, " "+kb.Help().Key)
			descriptions = append(descriptions, kb.Help().Desc+" ")
		}

		col := lipgloss.JoinHorizontal(lipgloss.Top,
			m.Styles.Key.Render(strings.Join(keys, "\n")),
			m.Styles.Key.Render(" "),
			m.Styles.Desc.Render(strings.Join(descriptions, "\n")),
		)

		colHeader := m.Styles.Header.Width(lipgloss.Width(col)).Render(group.Category)

		col = lipgloss.JoinVertical(lipgloss.Left, colHeader, col)

		outs = append(outs, m.Styles.Group.Render(col))
	}

	var cols []string
	var col string
	var colCount int
	for _, out := range outs {
		if lipgloss.Height(col)+lipgloss.Height(out) > height || colCount >= 3 {
			cols = append(cols, col)
			col = ""
			colCount = 0
		}

		col = lipgloss.JoinVertical(lipgloss.Left, col, out)
		colCount++
	}

	if col != "" {
		cols = append(cols, col)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func shouldRenderColumn(b KeyMapCategory) (ok bool) {
	for _, v := range b.Keys {
		if v.Enabled() {
			return true
		}
	}
	return false
}
