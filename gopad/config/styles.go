package config

import (
	"fmt"
	"image/color"

	"github.com/charmbracelet/lipgloss/v2"

	"go.gopad.dev/gopad/internal/bubbles"
	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/label"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

type ThemeStyles struct {
	Name       string
	Foreground color.Color
	Background color.Color

	Colors     bubbles.ColorStyles
	Icons      IconStyles
	UI         UiStyles
	Diagnostic DiagnosticStyles
	CodeStyles *CodeStyles
}

type CodeStyles struct {
	styles map[string]lipgloss.Style
	scopes []string
}

func (t *CodeStyles) Highlight(i int, languageName string) lipgloss.Style {
	scope := t.scopes[i]
	style, ok := t.styles[fmt.Sprintf("%s.%s", scope, languageName)]
	if ok {
		return style
	}
	return t.styles[scope]
}

func (t *CodeStyles) Scope(i int) string {
	return t.scopes[i]
}

func (t *CodeStyles) Scopes() []string {
	return t.scopes
}

type IconStyles struct {
	RootDir lipgloss.Style
	Dir     lipgloss.Style
	OpenDir lipgloss.Style
	File    lipgloss.Style

	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
	Hint    lipgloss.Style

	Files map[string]lipgloss.Style

	UnknownType lipgloss.Style
	Types       map[string]lipgloss.Style
}

func (c IconStyles) FileIcon(name string) lipgloss.Style {
	if r, ok := c.Files[name]; ok {
		return r
	}
	return c.File
}

func (c IconStyles) TypeIcon(name string) lipgloss.Style {
	if r, ok := c.Types[name]; ok {
		return r
	}
	return c.UnknownType
}

type UiStyles struct {
	AppBar  AppBarStyles
	CodeBar CodeBarStyles

	FileTree  FileTreeStyles
	FileView  FileViewStyles
	SearchBar SearchBarStyles

	Documentation DocumentationStyles
	Autocomplete  AutocompleteStyles
	Overlay       OverlayStyles

	TextInput         textinput.Styles
	Button            button.Styles
	FilePicker        filepicker.Styles
	Cursor            cursor.Styles
	Help              help.Styles
	NotificationStyle notifications.Styles
	List              list.Styles
	Label             label.Styles
}

type AppBarStyles struct {
	Style      lipgloss.Style
	TitleStyle lipgloss.Style

	Files AppBarFilesStyle
}

type AppBarFilesStyle struct {
	Style             lipgloss.Style
	FileStyle         lipgloss.Style
	SelectedFileStyle lipgloss.Style
}

type FileTreeStyles struct {
	Style                       lipgloss.Style
	EmptyStyle                  lipgloss.Style
	EntryStyle                  lipgloss.Style
	EntrySelectedStyle          lipgloss.Style
	EntrySelectedUnfocusedStyle lipgloss.Style
}

type SearchBarStyles struct {
	Style       lipgloss.Style
	ResultStyle lipgloss.Style
}

type FileViewStyles struct {
	Style       lipgloss.Style
	EmptyStyle  lipgloss.Style
	BorderStyle lipgloss.Style

	LineStyle       lipgloss.Style
	LinePrefixStyle lipgloss.Style
	LineCharStyle   lipgloss.Style

	CurrentLineStyle       lipgloss.Style
	CurrentLinePrefixStyle lipgloss.Style
	CurrentLineCharStyle   lipgloss.Style

	SelectionStyle lipgloss.Style
	InlayHintStyle lipgloss.Style
}

type CodeBarStyles struct {
	Style lipgloss.Style
}

type EditorStyles struct {
	EmptyStyle lipgloss.Style

	FileStyle         lipgloss.Style
	FileSelectedStyle lipgloss.Style

	CodeBorderStyle lipgloss.Style

	CodeLineStyle     lipgloss.Style
	CodePrefixStyle   lipgloss.Style
	CodeLineCharStyle lipgloss.Style

	CodeCurrentLineStyle       lipgloss.Style
	CodeCurrentLinePrefixStyle lipgloss.Style
	CodeCurrentLineCharStyle   lipgloss.Style

	CodeSelectionStyle lipgloss.Style
	CodeBarStyle       lipgloss.Style
	CodeInlayHintStyle lipgloss.Style

	FileTree      FileTreeStyles
	SearchBar     SearchBarStyles
	List          list.Styles
	Diagnostics   DiagnosticStyles
	Autocomplete  AutocompleteStyles
	Documentation DocumentationStyles

	CodeStyles map[string]lipgloss.Style
}

type DiagnosticStyles struct {
	ErrorStyle     lipgloss.Style
	ErrorCharStyle lipgloss.Style

	WarningStyle     lipgloss.Style
	WarningCharStyle lipgloss.Style

	InfoStyle     lipgloss.Style
	InfoCharStyle lipgloss.Style

	HintStyle     lipgloss.Style
	HintCharStyle lipgloss.Style

	DeprecatedStyle     lipgloss.Style
	DeprecatedCharStyle lipgloss.Style

	UnnecessaryStyle     lipgloss.Style
	UnnecessaryCharStyle lipgloss.Style
}

type DocumentationStyles struct {
	Style lipgloss.Style
}

type AutocompleteStyles struct {
	Style lipgloss.Style

	ItemStyle         lipgloss.Style
	SelectedItemStyle lipgloss.Style
}

type OverlayStyles struct {
	Styles          overlay.Styles
	RunOverlayStyle lipgloss.Style
}
