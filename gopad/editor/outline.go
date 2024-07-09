package editor

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

func Outline(f *file.File) tea.Cmd {
	return func() tea.Msg {
		return outlineMsg(f.OutlineTree())
	}
}

type outlineMsg []file.OutlineItem

type outlineItem struct {
	r        buffer.Range
	title    string
	rawTitle string
}

func (o outlineItem) Title() string {
	return o.title
}

func (o outlineItem) Description() string {
	return ""
}

func (o outlineItem) FilterValue() string {
	return o.rawTitle
}

func renderOutlineItem(file *file.File, itemStyle lipgloss.Style, item file.OutlineItem) outlineItem {
	codeCharStyle := config.Theme.UI.FileView.LineCharStyle

	var (
		title    string
		rawTitle string
	)
	for _, char := range item.Text {
		if char.Pos == nil {
			title += codeCharStyle.Copy().Inherit(itemStyle).Render(char.Char)
			rawTitle += char.Char
			continue
		}

		style := file.HighestMatchStyle(codeCharStyle, char.Pos.Row, char.Pos.Col)
		style = style.Copy().Inherit(itemStyle)

		title += style.Render(char.Char)
		rawTitle += char.Char
	}

	return outlineItem{
		r:        item.Range,
		title:    title,
		rawTitle: rawTitle,
	}
}

const OutlineOverlayID = "editor.outline"

var _ overlay.Overlay = (*OutlineOverlay)(nil)

func NewOutlineOverlay(f *file.File) OutlineOverlay {
	l := config.NewList[outlineItem](nil)
	l.TextInput.Placeholder = "Search symbols..."
	l.Focus()

	return OutlineOverlay{
		f: f,
		l: l,
	}
}

type OutlineOverlay struct {
	f     *file.File
	items []file.OutlineItem
	l     list.Model[outlineItem]
}

func (o OutlineOverlay) ID() string {
	return OutlineOverlayID
}

func (o OutlineOverlay) Position() (lipgloss.Position, lipgloss.Position) {
	return lipgloss.Center, lipgloss.Top
}

func (o OutlineOverlay) Margin() (int, int) {
	return 0, 2
}

func (o OutlineOverlay) Title() string {
	return "Outline"
}

func (o *OutlineOverlay) renderOutlineItems() {
	var out []outlineItem
	for i, item := range o.items {
		style := o.l.Styles.ItemStyle
		if i == o.l.SelectedIndex() {
			style = o.l.Styles.ItemSelectedStyle
		}
		out = append(out, renderOutlineItem(o.f, style, item))
	}
	o.l.SetItems(out)
}

func (o OutlineOverlay) Init() tea.Cmd {
	return tea.Sequence(
		textinput.Blink,
		Outline(o.f),
	)
}

func (o OutlineOverlay) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case outlineMsg:
		o.items = msg
		o.renderOutlineItems()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Cancel):
			return o, overlay.Close(OutlineOverlayID)
		case key.Matches(msg, config.Keys.OK):
			item := o.l.Selected()
			return o, tea.Batch(
				overlay.Close(OutlineOverlayID),
				file.Scroll(item.r.Start.Row, item.r.Start.Col),
			)
		}
	}

	var cmd tea.Cmd
	o.l, cmd = o.l.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	o.renderOutlineItems()

	if o.l.Clicked() {
		item := o.l.Selected()
		return o, tea.Batch(
			overlay.Close(OutlineOverlayID),
			file.Scroll(item.r.Start.Row, item.r.Start.Col),
		)
	}

	return o, tea.Batch(cmds...)
}

func (o OutlineOverlay) View(width int, height int) string {
	style := config.Theme.UI.Overlay.RunOverlayStyle
	width /= 2
	width -= style.GetHorizontalFrameSize()
	if width > 0 {
		o.l.SetWidth(width)
	}

	o.l.SetHeight(height - style.GetVerticalFrameSize() - 2)
	return o.l.View()
}
