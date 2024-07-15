package config

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/filetree"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/searchbar"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

func NewList[T list.Item](items []T) list.Model[T] {
	l := list.New(items)
	l.Styles = Theme.UI.List
	l.TextInput = NewTextInput()
	l.TextInput.Cursor = NewCursor()
	l.KeyMap = Keys.List()
	return l
}

func NewOverlays() overlay.Model {
	o := overlay.New()
	o.Styles = Theme.UI.Overlay.Styles

	return o
}

func NewTextInput() textinput.Model {
	ti := textinput.New()
	ti.Styles = Theme.UI.TextInput
	ti.Cursor = NewCursor()
	ti.KeyMap = Keys.Editor.TextInputKeyMap()
	return ti
}

func NewFilePicker(dir string, fileAllowed bool, dirAllowed bool) filepicker.Model {
	fp := filepicker.New()
	fp.ShowHidden = true
	fp.FileAllowed = fileAllowed
	fp.DirAllowed = dirAllowed
	fp.CurrentDirectory = dir
	fp.Styles = Theme.UI.FilePicker
	fp.KeyMap = Keys.FilePicker
	return fp
}

func NewButton(label string, onClick func() tea.Cmd) button.Model {
	b := button.New(label, onClick)
	b.Styles = Theme.UI.Button
	b.KeyMap = Keys.ButtonKeyMap()
	return b
}

func NewCursor() cursor.Model {
	c := cursor.New()
	c.Styles = Theme.UI.Cursor
	c.BlinkInterval = time.Duration(Gopad.Editor.Cursor.BlinkInterval)
	c.SetMode(Gopad.Editor.Cursor.Mode)
	c.Shape = Gopad.Editor.Cursor.Shape
	return c
}

func NewHelp() help.Model {
	h := help.New()
	h.Styles = Theme.UI.Help
	return h
}

func NewNotifications() notifications.Model {
	n := notifications.New()
	n.Styles = Theme.UI.NotificationStyle
	n.Margin = 1
	return n
}

func NewFileTree(openFile func(name string) tea.Cmd, languageIconFunc func(name string) lipgloss.Style) filetree.Model {
	ft := filetree.New()
	ft.Styles = Theme.UI.FileTree
	ft.KeyMap = Keys.Editor.FileTree
	ft.EmptyText = fmt.Sprintf("No folder open.\n\nPress '%s' to open a folder.", Keys.Editor.File.OpenFolder.Help().Key)
	ft.OpenFile = openFile
	ft.Icons = filetree.Icons{
		RootDir:          Theme.Icons.RootDir,
		Dir:              Theme.Icons.Dir,
		OpenDir:          Theme.Icons.OpenDir,
		File:             Theme.Icons.File,
		LanguageIconFunc: languageIconFunc,
	}
	ft.Ignored = Gopad.FileTree.Ignored

	return ft
}

func NewSearchBar(onSelect func(result searchbar.Result) tea.Cmd, onBlur tea.Cmd) searchbar.Model {
	sb := searchbar.New(onSelect, onBlur)
	sb.TextInput = NewTextInput()
	sb.TextInput.Placeholder = "type to search"
	sb.TextInput.Width = 20
	sb.Styles = Theme.UI.SearchBar
	sb.KeyMap = Keys.Editor.SearchBar

	return sb
}
