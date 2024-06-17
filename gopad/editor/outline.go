package editor

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

func Outline(f *File) tea.Cmd {
	return func() tea.Msg {
		if f.tree == nil {
			return nil
		}

		return outlineMsg(f.OutlineTree())
	}
}

type outlineMsg []OutlineItem

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

func renderOutlineItem(file *File, selectedStyle lipgloss.Style, selected bool, item OutlineItem) outlineItem {
	codeCharStyle := config.Theme.Editor.CodeLineCharStyle

	var (
		title    string
		rawTitle string
	)
	for _, char := range item.Text {
		if char.Pos == nil {
			style := codeCharStyle
			if selected {
				style = style.Reverse(true)
			}
			title += style.Render(char.Char)
			rawTitle += char.Char
			continue
		}

		style := file.HighestMatchStyle(codeCharStyle, char.Pos.Row, char.Pos.Col)
		if selected {
			style = selectedStyle.Inline(true).Copy().Inherit(style)
		}

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

func NewOutlineOverlay(f *File) OutlineOverlay {
	l := config.NewList[outlineItem](nil)
	l.TextInput.Placeholder = "Search symbols..."
	l.Focus()

	return OutlineOverlay{
		f:    f,
		list: l,
	}
}

type OutlineOverlay struct {
	f     *File
	items []OutlineItem
	list  list.Model[outlineItem]
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
		out = append(out, renderOutlineItem(o.f, o.list.Styles.ItemSelectedStyle, o.list.SelectedIndex() == i, item))
	}
	o.list.SetItems(out)
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
		case key.Matches(msg, config.Keys.OK):
			item := o.list.Selected().(outlineItem)
			return o, tea.Batch(
				overlay.Close(OutlineOverlayID),
				Scroll(item.r.Start.Row, item.r.Start.Col),
			)

		case key.Matches(msg, config.Keys.Cancel):
			return o, overlay.Close(OutlineOverlayID)
		}
	}

	var cmd tea.Cmd
	o.list, cmd = o.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	o.renderOutlineItems()

	return o, tea.Batch(cmds...)
}

func (o OutlineOverlay) View(width int, height int) string {
	style := config.Theme.Overlay.RunOverlayStyle
	width /= 2
	width -= style.GetHorizontalFrameSize()
	if width > 0 {
		o.list.SetWidth(width)
	}

	o.list.SetHeight(height - style.GetVerticalFrameSize() - 2)
	return o.list.View()
}
