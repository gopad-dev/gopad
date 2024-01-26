package editor

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
)

func Save() tea.Msg {
	return saveMsg{}
}

type saveMsg struct{}

func SaveAll() tea.Msg {
	return saveAllMsg{}
}

type saveAllMsg struct{}

func Close() tea.Msg {
	return closeMsg{}
}

type closeMsg struct{}

func CloseAll() tea.Msg {
	return closeAllMsg{}
}

type closeAllMsg struct{}

func Rename() tea.Msg {
	return renameMsg{}
}

type renameMsg struct{}

func Delete() tea.Msg {
	return deleteMsg{}
}

type deleteMsg struct{}

func SetLanguage(lang string) tea.Cmd {
	return func() tea.Msg {
		return setLanguageMsg{
			Language: lang,
		}
	}
}

type setLanguageMsg struct {
	Language string
}

func Paste() tea.Msg {
	text, err := clipboard.ReadAll()
	if err != nil {
		return notifications.Add(fmt.Sprintf("Error pasting: %s", err))()
	}
	return pasteMsg(text)
}

type pasteMsg []byte

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
			return cutMsg(s)
		}, notifications.Add("Cut to clipboard"))()
	}
}

type cutMsg buffer.Range

func Select(fromRow int, fromCol int, toRow int, toCol int) tea.Cmd {
	return func() tea.Msg {
		return selectMsg{
			FromRow: fromRow,
			FromCol: fromCol,
			ToRow:   toRow,
			ToCol:   toCol,
		}
	}
}

type selectMsg struct {
	FromRow int
	FromCol int
	ToRow   int
	ToCol   int
}

func Scroll(row int, col int) tea.Cmd {
	return func() tea.Msg {
		return scrollMsg{
			Row: row,
			Col: col,
		}
	}
}

type scrollMsg struct {
	Row int
	Col int
}

func OpenDir(name string) tea.Cmd {
	return func() tea.Msg {
		return openDirMsg{
			Name: name,
		}
	}
}

type openDirMsg struct {
	Name string
}

func OpenFile(name string) tea.Cmd {
	return func() tea.Msg {
		return openFileMsg{
			Name: name,
		}
	}
}

type openFileMsg struct {
	Name string
}

func SaveFile(name string) tea.Cmd {
	return func() tea.Msg {
		return saveFileMsg{
			Name: name,
		}
	}
}

type saveFileMsg struct {
	Name string
}

func CloseFile(name string) tea.Cmd {
	return func() tea.Msg {
		return closeFileMsg{
			Name: name,
		}
	}
}

type closeFileMsg struct {
	Name string
}

func FocusFile(name string) tea.Cmd {
	return func() tea.Msg {
		return focusFileMsg{
			Name: name,
		}
	}
}

type focusFileMsg struct {
	Name string
}

func NewFile(name string) tea.Cmd {
	return func() tea.Msg {
		return newFileMsg{
			Name: name,
		}
	}
}

type newFileMsg struct {
	Name string
}

func RenameFile(name string) tea.Cmd {
	return func() tea.Msg {
		return renameFileMsg{
			Name: name,
		}
	}
}

type renameFileMsg struct {
	Name string
}
