package config

import (
	"time"

	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
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
