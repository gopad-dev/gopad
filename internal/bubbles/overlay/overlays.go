package overlay

import (
	"slices"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Style        lipgloss.Style
	TitleStyle   lipgloss.Style
	ContentStyle lipgloss.Style
}

var DefaultStyles = Styles{
	Style:        lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.ANSIColor(13)).Padding(0, 1),
	TitleStyle:   lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(13)).Reverse(true),
	ContentStyle: lipgloss.NewStyle().Padding(0, 1, 1, 1),
}

func New() Model {
	return Model{
		Styles: DefaultStyles,
	}
}

type Model struct {
	overlays []Overlay

	Styles Styles
}

func (m *Model) Focused() bool {
	return len(m.overlays) > 0
}

func (m *Model) Has(id string) bool {
	return slices.IndexFunc(m.overlays, func(o Overlay) bool {
		return id == o.ID()
	}) != -1
}

func (m *Model) add(overlay Overlay) {
	i := slices.IndexFunc(m.overlays, func(o Overlay) bool {
		return overlay.ID() == o.ID()
	})
	if i != -1 {
		return
	}
	m.overlays = append(m.overlays, overlay)
}

func (m *Model) remove(id string) {
	m.overlays = slices.DeleteFunc(m.overlays, func(o Overlay) bool {
		return o.ID() == id
	})
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case openMsg:
		overlay, cmd := msg.overlay.Init()
		m.add(overlay)
		return m, cmd
	case closeMsg:
		m.remove(msg.id)
		return m, nil
	}

	if len(m.overlays) == 0 {
		return m, nil
	}

	overlayIndex := len(m.overlays) - 1
	var cmd tea.Cmd
	m.overlays[overlayIndex], cmd = m.overlays[overlayIndex].Update(msg)
	return m, cmd
}

func (m Model) View(width int, height int, background string) string {
	for _, overlay := range m.overlays {
		x, y := overlay.Position()
		marginX, marginY := overlay.Margin()

		maxOverlayWidth := width - marginX*2 - m.Styles.Style.GetHorizontalFrameSize() - m.Styles.ContentStyle.GetHorizontalFrameSize()
		maxOverlayHeight := height - marginY*2 - m.Styles.Style.GetVerticalFrameSize() - m.Styles.ContentStyle.GetVerticalFrameSize()
		if overlay.Title() != "" {
			maxOverlayHeight -= m.Styles.TitleStyle.GetVerticalFrameSize() + 1
		}

		background = PlacePosition(x, y,
			m.renderWithTitle(overlay.Title(), overlay.View(maxOverlayWidth, maxOverlayHeight)),
			background,
			WithMarginX(marginX),
			WithMarginY(marginY),
		)
	}
	return background
}

func (m Model) renderWithTitle(title string, content string) string {
	if title == "" {
		return m.Styles.Style.Render(content)
	}

	renderedContent := m.Styles.ContentStyle.Render(content)
	return m.Styles.Style.Render(lipgloss.JoinVertical(lipgloss.Center,
		m.Styles.TitleStyle.Width(lipgloss.Width(renderedContent)).Render(title),
		renderedContent,
	))
}
