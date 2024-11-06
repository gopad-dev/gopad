package gopad

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/editor"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

var Actions = []Action{
	{
		Name: "Quit",
		Run: func() tea.Cmd {
			return Quit
		},
	},
	{
		Name: "Help",
		Run: func() tea.Cmd {
			return overlay.Open(NewHelpOverlay())
		},
	},
	{
		Name: "Run",
		Run:  Terminal,
	},
	{
		Name: "New File",
		Run: func() tea.Cmd {
			return overlay.Open(editor.NewNewOverlay())
		},
	},
	{
		Name: "Open Folder",
		Run: func() tea.Cmd {
			path, err := os.Getwd()
			if err != nil {
				return notifications.Add(fmt.Sprintf("Error getting current working directory: %v", err))
			}
			return overlay.Open(editor.NewOpenOverlay(path, false, true))
		},
	},
	{
		Name: "Open File",
		Run: func() tea.Cmd {
			path, err := os.Getwd()
			if err != nil {
				return notifications.Add(fmt.Sprintf("Error getting current working directory: %v", err))
			}
			return overlay.Open(editor.NewOpenOverlay(path, true, false))
		},
	},
	{
		Name: "Save File",
		Run: func() tea.Cmd {
			return editor.SaveAction
		},
	},
	{
		Name: "Save all Files",
		Run: func() tea.Cmd {
			return editor.SaveAllAction
		},
	},
	{
		Name: "Rename File",
		Run: func() tea.Cmd {
			return editor.RenameAction
		},
	},
	{
		Name: "Close File",
		Run: func() tea.Cmd {
			return editor.CloseAction
		},
	},
	{
		Name: "Close All Files",
		Run: func() tea.Cmd {
			return editor.CloseAllAction
		},
	},
	{
		Name: "Delete File",
		Run: func() tea.Cmd {
			return editor.DeleteAction
		},
	},
	{
		Name: "Go To",
		Run: func() tea.Cmd {
			return editor.GoToAction
		},
	},
	{
		Name: "Set Language",
		Run: func() tea.Cmd {
			return overlay.Open(editor.NewSetLanguageOverlay())
		},
	},
	{
		Name: "Set Theme",
		Run: func() tea.Cmd {
			return overlay.Open(NewSetThemeOverlay())
		},
	},
	{
		Name: "Open Key Mapper",
		Run: func() tea.Cmd {
			return overlay.Open(NewKeyMapperOverlay())
		},
	},
	{
		Name: "Paste",
		Run: func() tea.Cmd {
			return editor.Paste
		},
	},
	{
		Name: "Start LSP",
		Run: func() tea.Cmd {
			return OpenLSPOverlay
		},
	},
}

type Action struct {
	Name string
	Run  func() tea.Cmd
}

func (a Action) Title() string {
	return a.Name
}

func (a Action) Description() string {
	return ""
}
