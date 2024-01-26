package gopad

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"

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
			return editor.Save
		},
	},
	{
		Name: "Save all Files",
		Run: func() tea.Cmd {
			return editor.SaveAll
		},
	},
	{
		Name: "Rename File",
		Run: func() tea.Cmd {
			return editor.Rename
		},
	},
	{
		Name: "Close File",
		Run: func() tea.Cmd {
			return editor.Close
		},
	},
	{
		Name: "Close All Files",
		Run: func() tea.Cmd {
			return editor.CloseAll
		},
	},
	{
		Name: "Delete File",
		Run: func() tea.Cmd {
			return overlay.Open(editor.NewDeleteOverlay())
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
