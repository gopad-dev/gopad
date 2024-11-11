package editor

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/editor/buffer"

	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

func SaveAction() tea.Msg {
	return SaveActionMsg{}
}

type SaveActionMsg struct{}

func SaveAllAction() tea.Msg {
	return SaveAllActionMsg{}
}

type SaveAllActionMsg struct{}

func CloseAction() tea.Msg {
	return CloseActionMsg{}
}

type CloseActionMsg struct{}

func CloseAllAction() tea.Msg {
	return CloseAllActionMsg{}
}

type CloseAllActionMsg struct{}

func RenameAction() tea.Msg {
	return RenameActionMsg{}
}

type RenameActionMsg struct{}

func DeleteAction() tea.Msg {
	return DeleteActionMsg{}
}

type DeleteActionMsg struct{}

func GoToAction() tea.Msg {
	return GoToActionMsg{}
}

type GoToActionMsg struct{}

func SetLanguageAction(lang string) tea.Cmd {
	return func() tea.Msg {
		return SetLanguageActionMsg{
			Language: lang,
		}
	}
}

type SetLanguageActionMsg struct {
	Language string
}

func Paste() tea.Msg {
	text, err := clipboard.ReadAll()
	if err != nil {
		return notifications.Add(fmt.Sprintf("Error pasting: %s", err))()
	}
	return tea.Batch(func() tea.Msg {
		return tea.PasteMsg(text)
	}, notifications.Add("Pasted from clipboard"))()
}

func Copy(b []byte) tea.Cmd {
	return func() tea.Msg {
		if err := clipboard.WriteAll(string(b)); err != nil {
			return notifications.Add(fmt.Sprintf("Error copying: %s", err))()
		}

		return notifications.Add("Copied to clipboard")()
	}
}

func Cut(s buffer.Range, b []byte) tea.Cmd {
	return func() tea.Msg {
		if err := clipboard.WriteAll(string(b)); err != nil {
			return notifications.Add(fmt.Sprintf("Error copying: %s", err))()
		}

		return tea.Batch(func() tea.Msg {
			return CutMsg(s)
		}, notifications.Add("Cut to clipboard"))()
	}
}

type CutMsg buffer.Range

func SelectAction(r buffer.Range) tea.Cmd {
	return func() tea.Msg {
		return SelectActionMsg(r)
	}
}

type SelectActionMsg buffer.Range

func ScrollAction(p buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return ScrollActionMsg(p)
	}
}

type ScrollActionMsg buffer.Point

func OpenDir(name string) tea.Cmd {
	return func() tea.Msg {
		return OpenDirMsg{
			Name: name,
		}
	}
}

type OpenDirMsg struct {
	Name string
}

func OpenFile(name string) tea.Cmd {
	return func() tea.Msg {
		return OpenFileMsg{
			Name: name,
		}
	}
}

func OpenFilePosition(name string, position *buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return OpenFileMsg{
			Name:     name,
			Position: position,
		}
	}
}

type OpenFileMsg struct {
	Name     string
	Position *buffer.Point
}

func SaveFile(name string) tea.Cmd {
	return func() tea.Msg {
		return SaveFileMsg{
			Name: name,
		}
	}
}

type SaveFileMsg struct {
	Name string
}

func CloseFile(name string) tea.Cmd {
	return func() tea.Msg {
		return CloseFileMsg{
			Name: name,
		}
	}
}

type CloseFileMsg struct {
	Name string
}

func NewFile(name string) tea.Cmd {
	return func() tea.Msg {
		return NewFileMsg{
			Name: name,
		}
	}
}

type NewFileMsg struct {
	Name string
}

func RenameFile(oldName string, newName string) tea.Cmd {
	return func() tea.Msg {
		return RenameFileMsg{
			OldName: oldName,
			NewName: newName,
		}
	}
}

type RenameFileMsg struct {
	OldName string
	NewName string
}

func DeleteFile(name string) tea.Cmd {
	return func() tea.Msg {
		return DeleteFileMsg{
			Name: name,
		}
	}
}

type DeleteFileMsg struct {
	Name string
}
