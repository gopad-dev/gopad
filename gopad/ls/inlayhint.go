package ls

import (
	"github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
)

func GetInlayHint(name string, r buffer.Range) tea.Cmd {
	return func() tea.Msg {
		return GetInlayHintMsg{
			Name:  name,
			Range: r,
		}
	}
}

type GetInlayHintMsg struct {
	Name  string
	Range buffer.Range
}

func UpdateInlayHint(name string, hints []InlayHint) tea.Cmd {
	return func() tea.Msg {
		return UpdateInlayHintMsg{
			Name:  name,
			Hints: hints,
		}
	}
}

type UpdateInlayHintMsg struct {
	Name  string
	Hints []InlayHint
}

func RefreshInlayHint() tea.Cmd {
	return func() tea.Msg {
		return RefreshInlayHintMsg{}
	}
}

type RefreshInlayHintMsg struct{}

type InlayHintType int

func (t InlayHintType) String() string {
	switch t {
	case InlayHintTypeNone:
		return "none"
	case InlayHintTypeType:
		return "type"
	case InlayHintTypeParameter:
		return "parameter"
	default:
		return "unknown"
	}
}

const (
	InlayHintTypeNone      InlayHintType = 0
	InlayHintTypeType      InlayHintType = 1
	InlayHintTypeParameter InlayHintType = 2
)

type InlayHint struct {
	Type         InlayHintType
	Position     buffer.Position
	Label        string
	Tooltip      string
	PaddingLeft  bool
	PaddingRight bool
}
