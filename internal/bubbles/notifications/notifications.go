package notifications

import (
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

func Add(content string) tea.Cmd {
	return func() tea.Msg {
		return addMsg{
			Content: content,
		}
	}
}

func Addf(format string, a ...any) tea.Cmd {
	return Add(fmt.Sprintf(format, a...))
}

func Remove(id int) tea.Cmd {
	return func() tea.Msg {
		return removeMsg{
			ID: id,
		}
	}
}

type (
	addMsg struct {
		Content string
	}
	removeMsg struct {
		ID int
	}
)

type notification struct {
	ID      int
	Content string
}

type Styles struct {
	Notification lipgloss.Style
}

var DefaultStyles = Styles{
	Notification: lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder()),
}

func New() Model {
	return Model{
		Styles:  DefaultStyles,
		Margin:  0,
		Timeout: 3 * time.Second,
	}
}

type Model struct {
	notifications []notification
	lastID        int
	bg            string

	Styles  Styles
	Margin  int
	Timeout time.Duration
}

func (m *Model) SetBackground(bg string) {
	m.bg = bg
}

func (m *Model) Active() bool {
	return len(m.notifications) > 0
}

func (m *Model) add(content string) tea.Cmd {
	m.lastID++
	id := m.lastID
	log.Printf("Adding notification with ID %d: %s\n", id, content)
	m.notifications = append(m.notifications, notification{
		ID:      id,
		Content: content,
	})

	return tea.Tick(m.Timeout, func(_ time.Time) tea.Msg {
		return removeMsg{
			ID: id,
		}
	})
}

func (m *Model) remove(id int) {
	m.notifications = slices.DeleteFunc(m.notifications, func(n notification) bool {
		return n.ID == id
	})
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case addMsg:
		if cmd := m.add(msg.Content); cmd != nil {
			cmds = append(cmds, cmd)
		}
	case removeMsg:
		m.remove(msg.ID)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View(_ int, height int) string {
	height = height - 1 - m.Margin

	var str string
	for i := len(m.notifications) - 1; i >= 0; i-- {
		content := m.Styles.Notification.Render(m.notifications[i].Content)
		contentHeight := lipgloss.Height(content)

		if height-contentHeight <= 0 {
			break
		}

		height -= contentHeight
		str = lipgloss.JoinVertical(lipgloss.Right, str, content)
	}
	if str == "" {
		return m.bg
	}

	return overlay.PlacePosition(lipgloss.Right, lipgloss.Bottom, str, m.bg, overlay.WithMarginY(m.Margin), overlay.WithMarginX(1))
}
