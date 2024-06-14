package ls

import (
	tea "github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
)

func GetAutocompletion(name string, row int, col int) tea.Cmd {
	return func() tea.Msg {
		return GetAutocompletionMsg{
			Name: name,
			Row:  row,
			Col:  col,
		}
	}
}

type GetAutocompletionMsg struct {
	Name string
	Row  int
	Col  int
}

func UpdateAutocompletion(name string, completions []CompletionItem) tea.Cmd {
	return func() tea.Msg {
		return UpdateAutocompletionMsg{
			Name:        name,
			Completions: completions,
		}
	}
}

type UpdateAutocompletionMsg struct {
	Name        string
	Completions []CompletionItem
	Selected    int
}

type CompletionItem struct {
	Label      string
	Detail     string
	Kind       CompletionItemKind
	Text       string
	Edit       *TextEdit
	Deprecated bool
}

type TextEdit struct {
	Range   buffer.Range
	NewText string
}

type CompletionItemKind int

func (i CompletionItemKind) String() string {
	switch i {
	case Text:
		return "Text"
	case Method:
		return "Method"
	case Function:
		return "Function"
	case Constructor:
		return "Constructor"
	case Field:
		return "Field"
	case Variable:
		return "Variable"
	case Class:
		return "Class"
	case Interface:
		return "Interface"
	case Module:
		return "Module"
	case Property:
		return "Property"
	case Unit:
		return "Unit"
	case Value:
		return "Value"
	case Enum:
		return "Enum"
	case Keyword:
		return "Keyword"
	case Snippet:
		return "Snippet"
	case Color:
		return "Color"
	case File:
		return "File"
	case Reference:
		return "Reference"
	case Folder:
		return "Folder"
	case EnumMember:
		return "EnumMember"
	case Constant:
		return "Constant"
	case Struct:
		return "Struct"
	case Event:
		return "Event"
	case Operator:
		return "Operator"
	case TypeParameter:
		return "TypeParameter"
	default:
		return "Unknown"
	}
}

const (
	Text CompletionItemKind = iota
	Method
	Function
	Constructor
	Field
	Variable
	Class
	Interface
	Module
	Property
	Unit
	Value
	Enum
	Keyword
	Snippet
	Color
	File
	Reference
	Folder
	EnumMember
	Constant
	Struct
	Event
	Operator
	TypeParameter
)
