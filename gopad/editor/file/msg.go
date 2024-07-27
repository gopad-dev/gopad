package file

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
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
	return PasteMsg(text)
}

type PasteMsg []byte

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

func Select(fromRow int, fromCol int, toRow int, toCol int) tea.Cmd {
	return func() tea.Msg {
		return SelectMsg{
			FromRow: fromRow,
			FromCol: fromCol,
			ToRow:   toRow,
			ToCol:   toCol,
		}
	}
}

type SelectMsg struct {
	FromRow int
	FromCol int
	ToRow   int
	ToCol   int
}

func Scroll(row int, col int) tea.Cmd {
	return func() tea.Msg {
		return ScrollMsg{
			Row: row,
			Col: col,
		}
	}
}

type ScrollMsg struct {
	Row int
	Col int
}

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

func OpenFilePosition(name string, position *buffer.Position) tea.Cmd {
	return func() tea.Msg {
		return OpenFileMsg{
			Name:     name,
			Position: position,
		}
	}
}

type OpenFileMsg struct {
	Name     string
	Position *buffer.Position
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

func FocusFile(name string) tea.Cmd {
	return func() tea.Msg {
		return FocusFileMsg{
			Name: name,
		}
	}
}

type FocusFileMsg struct {
	Name string
}

func BlurFile() tea.Msg {
	return BlurFileMsg{}
}

type BlurFileMsg struct{}

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

func RenameFile(name string) tea.Cmd {
	return func() tea.Msg {
		return RenameFileMsg{
			Name: name,
		}
	}
}

type RenameFileMsg struct {
	Name string
}

func GoTo() tea.Msg {
	return GoToMsg{}
}

type GoToMsg struct{}
