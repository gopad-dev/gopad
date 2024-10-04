package ls

import (
	"github.com/charmbracelet/bubbletea/v2"
)

func Err(err error) tea.Cmd {
	return func() tea.Msg {
		return err
	}
}

func WorkspaceOpened(workspace string) tea.Cmd {
	return func() tea.Msg {
		return WorkspaceOpenedMsg{
			Workspace: workspace,
		}
	}
}

type WorkspaceOpenedMsg struct {
	Workspace string
}

func WorkspaceClosed(workspace string) tea.Cmd {
	return func() tea.Msg {
		return WorkspaceClosedMsg{
			Workspace: workspace,
		}
	}
}

type WorkspaceClosedMsg struct {
	Workspace string
}

func FileCreated(name string, text []byte) tea.Cmd {
	return func() tea.Msg {
		return FileCreatedMsg{
			Name: name,
			Text: text,
		}
	}
}

type FileCreatedMsg struct {
	Name string
	Text []byte
}

func FileOpened(name string, version int32, text []byte) tea.Cmd {
	return func() tea.Msg {
		return FileOpenedMsg{
			Name:    name,
			Version: version,
			Text:    text,
		}
	}
}

type FileOpenedMsg struct {
	Name    string
	Version int32
	Text    []byte
}

func FileClosed(name string) tea.Cmd {
	return func() tea.Msg {
		return FileClosedMsg{
			Name: name,
		}
	}
}

type FileClosedMsg struct {
	Name string
}

func FileChanged(name string, version int32, text []byte) tea.Cmd {
	return func() tea.Msg {
		return FileChangedMsg{
			Name:    name,
			Version: version,
			Text:    text,
		}
	}
}

type FileChangedMsg struct {
	Name    string
	Version int32
	Text    []byte
}

func FileSaved(name string, text []byte) tea.Cmd {
	return func() tea.Msg {
		return FileSavedMsg{
			Name: name,
			Text: text,
		}
	}
}

type FileSavedMsg struct {
	Name string
	Text []byte
}

func FileRenamed(oldName string, newName string) tea.Cmd {
	return func() tea.Msg {
		return FileRenamedMsg{
			OldName: oldName,
			NewName: newName,
		}
	}
}

type FileRenamedMsg struct {
	OldName string
	NewName string
}

func FileDeleted(name string) tea.Cmd {
	return func() tea.Msg {
		return FileDeletedMsg{
			Name: name,
		}
	}
}

type FileDeletedMsg struct {
	Name string
}
