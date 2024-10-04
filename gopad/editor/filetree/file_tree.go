package filetree

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/bubblezone"
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/editormsg"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

const (
	zoneID       = "file_tree"
	zoneIDPrefix = "file_tree:"
)

func fileIconByFileNameFunc(name string) lipgloss.Style {
	language := file.GetLanguageByFilename(name)
	var languageName string
	if language != nil {
		languageName = language.Name
	}
	return config.Theme.Icons.FileIcon(languageName)
}

type Entry struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Entry
	Selected bool
	Open     bool
}

func (e *Entry) Sort() {
	slices.SortFunc(e.Children, func(a *Entry, b *Entry) int {
		if a.IsDir && !b.IsDir {
			return -1
		}
		if !a.IsDir && b.IsDir {
			return 1
		}
		return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	for _, entry := range e.Children {
		entry.Sort()
	}
}

func Refresh() tea.Msg {
	return refreshMsg{}
}

type refreshMsg struct{}

func New() Model {
	return Model{
		Width:     24,
		EmptyText: "No folder open",
	}
}

type Model struct {
	entry     *Entry
	focus     bool
	show      bool
	offset    int
	Width     int
	EmptyText string
	Ignored   []string
}

func (m *Model) Open(name string) error {
	root := &Entry{
		Name:     filepath.Base(name),
		Path:     name,
		IsDir:    true,
		Open:     true,
		Selected: true,
	}

	if err := filepath.WalkDir(name, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name == path {
			return nil
		}

		relPath, err := filepath.Rel(name, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		current := root
		for _, component := range strings.Split(relPath, string(os.PathSeparator)) {
			if slices.Contains(m.Ignored, component) {
				return filepath.SkipDir
			}
			found := false
			for _, child := range current.Children {
				if child.Name == component {
					current = child
					found = true
					break
				}
			}
			if !found {
				node := &Entry{
					Name:  component,
					Path:  path,
					IsDir: d.IsDir(),
				}
				current.Children = append(current.Children, node)
				current = node
			}
		}

		return nil
	}); err != nil {
		return err
	}

	root.Sort()

	m.entry = root
	return nil
}

func (m *Model) Visible() bool {
	return m.show
}

func (m *Model) Show() {
	m.show = true
}

func (m *Model) Hide() {
	m.show = false
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) selectIndex(i int) *Entry {
	if m.entry == nil {
		return nil
	}

	var currentIndex int
	var selected *Entry
	var walk func(*Entry)
	walk = func(e *Entry) {
		if currentIndex == i {
			e.Selected = true
			selected = e
		} else {
			e.Selected = false
		}

		currentIndex++
		if e.IsDir && !e.Open {
			return
		}

		for _, child := range e.Children {
			walk(child)
		}
	}
	walk(m.entry)
	return selected
}

func (m *Model) SelectNext() {
	if m.entry == nil {
		return
	}

	var walk func(*Entry) bool
	var prev *Entry
	var next *Entry
	walk = func(e *Entry) bool {
		if e.Selected {
			prev = e
		} else {
			if prev != nil {
				next = e
				return true
			}
		}

		if e.IsDir && !e.Open {
			return false
		}

		for _, child := range e.Children {
			if walk(child) {
				return true
			}
		}
		return false
	}
	walk(m.entry)

	if next != nil {
		next.Selected = true
		prev.Selected = false
	}
}

func (m *Model) SelectPrev() {
	if m.entry == nil {
		return
	}

	var walk func(*Entry)
	var prev *Entry
	walk = func(e *Entry) {
		if e.Selected {
			if prev != nil {
				prev.Selected = true
				e.Selected = false
				return
			}
		}

		prev = e

		if e.IsDir && !e.Open {
			return
		}

		for _, child := range e.Children {
			walk(child)
		}
	}
	walk(m.entry)
}

func (m *Model) Selected() *Entry {
	if m.entry == nil {
		return nil
	}

	var walk func(*Entry) *Entry
	walk = func(e *Entry) *Entry {
		if e.Selected {
			return e
		}
		for _, child := range e.Children {
			if entry := walk(child); entry != nil {
				return entry
			}
		}
		return nil
	}
	return walk(m.entry)
}

func (m Model) zoneEntryID(i int) string {
	return fmt.Sprintf("%s%d", zoneIDPrefix, i)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case refreshMsg:
		if err := m.Open(m.entry.Path); err != nil {
			cmds = append(cmds, notifications.Add("Error updating file tree: "+err.Error()))
		}
		return m, tea.Batch(cmds...)
	case tea.MouseClickMsg:
		for _, z := range zone.GetPrefix(zoneIDPrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				if !m.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFileTree))
				}

				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), zoneIDPrefix))
				m.selectIndex(i)

				return m, tea.Batch(cmds...)
			}
		}
	case tea.MouseReleaseMsg:
		for _, z := range zone.GetPrefix(zoneIDPrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				if !m.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFileTree))
				}

				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), zoneIDPrefix))
				entry := m.selectIndex(i)

				if entry.IsDir {
					entry.Open = !entry.Open
				} else {
					cmds = append(cmds, file.OpenFile(entry.Path))
				}

				return m, tea.Batch(cmds...)
			}
		}
	case tea.MouseWheelMsg:
		switch {
		case mouse.Matches(msg, zoneID, tea.MouseWheelUp):
			if !m.Focused() {
				cmds = append(cmds, editormsg.Focus(editormsg.ModelFileTree))
			}
			m.SelectPrev()
			return m, tea.Batch(cmds...)
		case mouse.Matches(msg, zoneID, tea.MouseWheelDown):
			if !m.Focused() {
				cmds = append(cmds, editormsg.Focus(editormsg.ModelFileTree))
			}
			m.SelectNext()
			return m, tea.Batch(cmds...)
		}
	case tea.KeyPressMsg:
		if m.Focused() {
			switch {
			case key.Matches(msg, config.Keys.Editor.FileTree.Refresh):
				cmds = append(cmds, Refresh)
			case key.Matches(msg, config.Keys.Editor.FileTree.Open):
				selected := m.Selected()
				if selected.IsDir {
					selected.Open = !selected.Open
				} else {
					cmds = append(cmds, file.OpenFile(selected.Path))
				}
			case key.Matches(msg, config.Keys.Editor.FileTree.SelectNext):
				m.SelectNext()
			case key.Matches(msg, config.Keys.Editor.FileTree.SelectPrev):
				m.SelectPrev()
			case key.Matches(msg, config.Keys.Editor.FileTree.ExpandWidth):
				m.Width++
			case key.Matches(msg, config.Keys.Editor.FileTree.ShrinkWidth):
				if m.Width > 6 {
					m.Width--
				}
			}
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) refreshViewOffset(selected int, height int) {
	if selected >= m.offset+height {
		m.offset = selected - height + 1
	} else if selected < m.offset {
		m.offset = selected
	}
}

func (m Model) View(height int) string {
	if m.entry == nil {
		return config.Theme.UI.FileTree.Style.Render(config.Theme.UI.FileTree.EmptyStyle.Height(height).Width(m.Width).Render(m.EmptyText))
	}

	var i int
	var entries []string
	var selected int
	var walk func(*Entry, string)
	walk = func(e *Entry, indent string) {
		if e.Selected {
			selected = len(entries)
		}
		entries = append(entries, m.entryView(e, i, indent))
		i++

		if e.IsDir && !e.Open {
			return
		}
		for _, child := range e.Children {
			walk(child, indent+"  ")
		}
	}
	walk(m.entry, "")

	m.refreshViewOffset(selected, height)

	var tree string
	for i := range height {
		ln := i + m.offset
		if ln >= len(entries) {
			tree += "\n"
			continue
		}

		entry := entries[ln]
		tree += entry + "\n"
	}

	tree = strings.TrimSuffix(tree, "\n")

	return zone.Mark(zoneID, config.Theme.UI.FileTree.Style.Height(height).Width(m.Width).Render(tree))
}

func (m Model) entryView(e *Entry, i int, indent string) string {
	var icon lipgloss.Style
	if e.IsDir {
		if indent == "" {
			icon = config.Theme.Icons.RootDir
		} else if e.Open {
			icon = config.Theme.Icons.OpenDir
		} else {
			icon = config.Theme.Icons.Dir
		}
	} else {
		icon = fileIconByFileNameFunc(e.Name)
		if icon.String() == "" {
			icon = config.Theme.Icons.File
		}
	}

	entryStyle := config.Theme.UI.FileTree.EntryStyle
	if e.Selected {
		if m.Focused() {
			entryStyle = config.Theme.UI.FileTree.EntrySelectedStyle
		} else {
			entryStyle = config.Theme.UI.FileTree.EntrySelectedUnfocusedStyle
		}
	}

	icon = entryStyle.Inherit(icon).SetString(icon.String())

	line := entryStyle.Render(indent) + icon.Render() + entryStyle.Render(" "+e.Name)
	if ansi.StringWidth(line) > m.Width {
		line = ansi.Truncate(line, m.Width, entryStyle.Render("…"))
	} else {
		line += entryStyle.Render(strings.Repeat(" ", m.Width-lipgloss.Width(line)))
	}

	return zone.Mark(m.zoneEntryID(i), line)
}
