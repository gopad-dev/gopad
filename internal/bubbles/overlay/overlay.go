package overlay

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Open(overlay Overlay) tea.Cmd {
	return func() tea.Msg {
		return openMsg{
			overlay: overlay,
		}
	}
}

func Close(id string) tea.Cmd {
	return tea.Batch(func() tea.Msg {
		return closeMsg{
			id: id,
		}
	}, ResetFocus)
}

func ResetFocus() tea.Msg {
	return ResetFocusMsg{}
}

type ResetFocusMsg struct{}

type openMsg struct {
	overlay Overlay
}

type closeMsg struct {
	id string
}

type Overlay interface {
	ID() string
	Position() (lipgloss.Position, lipgloss.Position)
	Margin() (int, int)
	Title() string

	Init() tea.Cmd
	Update(msg tea.Msg) (Overlay, tea.Cmd)
	View(width int, height int) string
}

func Render() {

}
