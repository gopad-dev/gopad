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
	// FullHelpView returns an extended group of help items, grouped by columns.
	// The help bubble will render the help in the order in which the help
	// items are returned here.
	FullHelpView() []KeyMapCategory
}

type KeyMapCategory struct {
	Category string
	Keys     []key.Binding
}

// Styles is a set of available style definitions for the Help bubble.
type Styles struct {
	Ellipsis lipgloss.Style

	Header lipgloss.Style

	// Styling for the short help
	ShortKey       lipgloss.Style
	ShortDesc      lipgloss.Style
	ShortSeparator lipgloss.Style

	// Styling for the full help
	FullKey       lipgloss.Style
	FullDesc      lipgloss.Style
	FullSeparator lipgloss.Style
}

var (
	keyStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})

	descStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#B2B2B2",
		Dark:  "#4A4A4A",
	})

	sepStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#DDDADA",
		Dark:  "#3C3C3C",
	})
)

var DefaultStyles = Styles{
	Header:         lipgloss.NewStyle().Reverse(true).AlignHorizontal(lipgloss.Center),
	ShortKey:       keyStyle,
	ShortDesc:      descStyle,
	ShortSeparator: sepStyle,
	Ellipsis:       sepStyle,
	FullKey:        keyStyle,
	FullDesc:       descStyle,
	FullSeparator:  sepStyle,
}

// Model contains the state of the help view.
type Model struct {
	ShortSeparator string
	FullSeparator  string

	// The symbol we use in the short help when help items have been truncated
	// due to width. Periods of ellipsis by default.
	Ellipsis string

	Styles Styles
}

// New creates a new help view with some useful defaults.
func New() Model {
	return Model{
		ShortSeparator: " • ",
		FullSeparator:  "    ",
		Ellipsis:       "…",
		Styles:         DefaultStyles,
	}
}

// Update helps satisfy the Bubble Tea Model interface. It's a no-op.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the help view's current state.
func (m Model) View(width int, height int, k KeyMap) string {
	return m.FullHelpView(width, height, k.FullHelpView())
}

// FullHelpView renders help columns from a slice of key binding slices. Each
// top level slice entry renders into a column.
func (m Model) FullHelpView(width int, height int, groups []KeyMapCategory) string {
	if len(groups) == 0 {
		return ""
	}

	// Linter note: at this time we don't think it's worth the additional
	// code complexity involved in preallocating this slice.
	//nolint:prealloc
	var (
		out []string

		totalWidth int
		sep        = m.Styles.FullSeparator.Render(m.FullSeparator)
		sepWidth   = lipgloss.Width(sep)
	)

	// Iterate over groups to build columns
	for i, group := range groups {
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
			keys = append(keys, kb.Help().Key)
			descriptions = append(descriptions, kb.Help().Desc)
		}

		col := lipgloss.JoinHorizontal(lipgloss.Top,
			m.Styles.FullKey.Render(strings.Join(keys, "\n")),
			m.Styles.FullKey.Render(" "),
			m.Styles.FullDesc.Render(strings.Join(descriptions, "\n")),
		)

		colWidth := lipgloss.Width(col)
		colHeader := m.Styles.Header.Width(colWidth).Render(group.Category)

		col = lipgloss.JoinVertical(lipgloss.Left, colHeader, col)

		// Column
		totalWidth += colWidth
		if width > 0 && totalWidth > width {
			break
		}

		out = append(out, col)

		// Separator
		if i < len(group.Keys)-1 {
			totalWidth += sepWidth
			if width > 0 && totalWidth > width {
				break
			}
		}

		out = append(out, sep)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}

func shouldRenderColumn(b KeyMapCategory) (ok bool) {
	for _, v := range b.Keys {
		if v.Enabled() {
			return true
		}
	}
	return false
}
