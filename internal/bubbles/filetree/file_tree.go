package filetree

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mattn/go-runewidth"

	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

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

type Styles struct {
	Style                       lipgloss.Style
	EmptyStyle                  lipgloss.Style
	EntryPrefixStyle            lipgloss.Style
	EntryStyle                  lipgloss.Style
	EntrySelectedStyle          lipgloss.Style
	EntrySelectedUnfocusedStyle lipgloss.Style
}

var DefaultStyles = Styles{
	Style:                       lipgloss.NewStyle().MarginRight(1),
	EmptyStyle:                  lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center),
	EntryPrefixStyle:            lipgloss.NewStyle().Faint(true),
	EntryStyle:                  lipgloss.NewStyle(),
	EntrySelectedStyle:          lipgloss.NewStyle().Reverse(true),
	EntrySelectedUnfocusedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Reverse(true),
}

type KeyMap struct {
	SelectPrev  key.Binding
	SelectNext  key.Binding
	ExpandWidth key.Binding
	ShrinkWidth key.Binding
	Open        key.Binding
	Refresh     key.Binding
}

func (m KeyMap) FullHelpView() []help.KeyMapCategory {
	return []help.KeyMapCategory{
		{
			Category: "File Tree",
			Keys: []key.Binding{
				m.SelectPrev,
				m.SelectNext,
				m.ExpandWidth,
				m.ShrinkWidth,
				m.Open,
			},
		},
	}
}

var DefaultKeyMap = KeyMap{
	SelectPrev: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("up", "select previous entry"),
	),
	SelectNext: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("down", "select next entry"),
	),
	ExpandWidth: key.NewBinding(
		key.WithKeys("ctrl+right"),
		key.WithHelp("ctrl+right", "expand width"),
	),
	ShrinkWidth: key.NewBinding(
		key.WithKeys("ctrl+left"),
		key.WithHelp("ctrl+left", "shrink width"),
	),
	Open: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open file or directory"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "refresh file tree"),
	),
}

type Icons struct {
	RootDir          lipgloss.Style
	Dir              lipgloss.Style
	OpenDir          lipgloss.Style
	File             lipgloss.Style
	LanguageIconFunc func(string) lipgloss.Style
}

var DefaultIcons = Icons{
	RootDir: lipgloss.NewStyle().SetString("📁"),
	Dir:     lipgloss.NewStyle().SetString("'📂'"),
	OpenDir: lipgloss.NewStyle().SetString("📁"),
	File:    lipgloss.NewStyle().SetString("📄"),
}

func New() Model {
	return Model{
		Width:      24,
		Styles:     DefaultStyles,
		KeyMap:     DefaultKeyMap,
		Icons:      DefaultIcons,
		EmptyText:  "No folder open",
		zonePrefix: zone.NewPrefix(),
	}
}

type Model struct {
	entry      *Entry
	focus      bool
	show       bool
	offset     int
	Width      int
	Styles     Styles
	Icons      Icons
	KeyMap     KeyMap
	EmptyText  string
	Ignored    []string
	OpenFile   func(string) tea.Cmd
	zonePrefix string
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
	return fmt.Sprintf("file_tree:%s:%d", m.zonePrefix, i)
}

func (m Model) zoneID() string {
	return fmt.Sprintf("file_tree:%s", m.zonePrefix)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case refreshMsg:
		if err := m.Open(m.entry.Path); err != nil {
			cmds = append(cmds, notifications.Add("Error updating file tree: "+err.Error()))
		}
		return m, tea.Batch(cmds...)
	case tea.MouseMsg:
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
			zonePrefix := fmt.Sprintf("file_tree:%s:", m.zonePrefix)
			for _, z := range zone.GetPrefix(zonePrefix) {
				if z.InBounds(msg) {
					i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), zonePrefix))
					entry := m.selectIndex(i)

					if entry.IsDir {
						entry.Open = !entry.Open
					} else {
						cmds = append(cmds, m.OpenFile(entry.Path))
					}

					return m, tea.Batch(cmds...)
				}
			}
		}

		switch {
		case mouse.Matches(msg, m.zoneID(), tea.MouseButtonWheelUp):
			m.SelectPrev()
			return m, nil
		case mouse.Matches(msg, m.zoneID(), tea.MouseButtonWheelDown):
			m.SelectNext()
			return m, nil
		}
	case tea.KeyMsg:
		if m.focus {
			switch {
			case key.Matches(msg, m.KeyMap.Refresh):
				cmds = append(cmds, Refresh)
			case key.Matches(msg, m.KeyMap.Open):
				selected := m.Selected()
				if selected.IsDir {
					selected.Open = !selected.Open
				} else {
					cmds = append(cmds, m.OpenFile(selected.Path))
				}
			case key.Matches(msg, m.KeyMap.SelectNext):
				m.SelectNext()
			case key.Matches(msg, m.KeyMap.SelectPrev):
				m.SelectPrev()
			case key.Matches(msg, m.KeyMap.ExpandWidth):
				m.Width++
			case key.Matches(msg, m.KeyMap.ShrinkWidth):
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
		return m.Styles.Style.Render(m.Styles.EmptyStyle.Height(height).Width(m.Width).Render(m.EmptyText))
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

	return zone.Mark(m.zoneID(), m.Styles.Style.Height(height).Width(m.Width).Render(tree))
}

func (m Model) entryView(e *Entry, i int, indent string) string {
	var icon lipgloss.Style

	if e.IsDir {
		if indent == "" {
			icon = m.Icons.RootDir
		} else {
			if e.Open {
				icon = m.Icons.OpenDir
			} else {
				icon = m.Icons.Dir
			}
		}
	} else {
		if m.Icons.LanguageIconFunc != nil {
			icon = m.Icons.LanguageIconFunc(e.Name)
		}
		if icon.String() == "" {
			icon = m.Icons.File
		}
	}

	line := indent + icon.Render() + " " + e.Name
	if runewidth.StringWidth(line) > m.Width {
		line = runewidth.Truncate(line, m.Width, "…")
	} else {
		line += strings.Repeat(" ", m.Width-lipgloss.Width(line))
	}

	if e.Selected {
		if m.Focused() {
			line = m.Styles.EntrySelectedStyle.Render(line)
		} else {
			line = m.Styles.EntrySelectedUnfocusedStyle.Render(line)
		}
	} else {
		line = m.Styles.EntryStyle.Render(line)
	}

	return zone.Mark(m.zoneEntryID(i), line)
}
