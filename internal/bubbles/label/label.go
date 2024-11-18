package label

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

func DefaultStyles() Styles {
	return Styles{
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
	}
}

type Styles struct {
	Style lipgloss.Style
}

func New(text string) Model {
	return Model{
		Text:   text,
		Styles: DefaultStyles(),
	}
}

type Model struct {
	Text   string
	Styles Styles
}

func (m Model) UpdateText(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(view string) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.Styles.Style.Render(m.Text),
		view,
	)
}
