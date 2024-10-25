package file

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/buffer"
)

func Save() tea.Msg {
	return SaveMsg{}
}

type SaveMsg struct{}

func SaveAll() tea.Msg {
	return saveAllMsg{}
}

type saveAllMsg struct{}

func Close() tea.Msg {
	return CloseMsg{}
}

type CloseMsg struct{}

func CloseAll() tea.Msg {
	return CloseAllMsg{}
}

type CloseAllMsg struct{}

func Rename() tea.Msg {
	return RenameMsg{}
}

type RenameMsg struct{}

func Delete() tea.Msg {
	return DeleteMsg{}
}

type DeleteMsg struct{}

func SetLanguage(lang string) tea.Cmd {
	return func() tea.Msg {
		return SetLanguageMsg{
			Language: lang,
		}
	}
}

type SetLanguageMsg struct {
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

func Select(r buffer.Range) tea.Cmd {
	return func() tea.Msg {
		return SelectMsg(r)
	}
}

type SelectMsg buffer.Range

func Scroll(p buffer.Point) tea.Cmd {
	return func() tea.Msg {
		return ScrollMsg(p)
	}
}

type ScrollMsg buffer.Point

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

func GoTo() tea.Msg {
	return GoToMsg{}
}

type GoToMsg struct{}
