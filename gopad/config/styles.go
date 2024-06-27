package config

import (
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/textinput"

	"go.gopad.dev/gopad/internal/bubbles/filetree"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/searchbar"
)

type ThemeConfig struct {
	Name string

	Colors Colors
	Icons  IconsConfig

	AppBarStyle      lipgloss.Style
	AppBarTitleStyle lipgloss.Style

	Editor EditorStyles

	Overlay OverlayStyles

	TextInput         textinput.Styles
	Button            button.Styles
	FilePicker        filepicker.Styles
	Cursor            cursor.Styles
	Help              help.Styles
	NotificationStyle notifications.Styles
	List              list.Styles
}

type Colors struct {
	PrimaryColor         lipgloss.Color
	PrimarySelectedColor lipgloss.Color

	PrimaryTextColor   lipgloss.Color
	SecondaryTextColor lipgloss.Color
	DisabledTextColor  lipgloss.Color

	BackgroundColor          lipgloss.Color
	SecondaryBackgroundColor lipgloss.Color

	CursorColor         lipgloss.Color
	DisabledCursorColor lipgloss.Color
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

	FileTree      filetree.Styles
	SearchBar     searchbar.Styles
	List          list.Styles
	Diagnostics   DiagnosticStyles
	Autocomplete  AutocompleteStyles
	Documentation DocumentationStyles

	CodeStyles map[string]lipgloss.Style
}

type DiagnosticStyles struct {
	ErrorStyle           lipgloss.Style
	ErrorCharStyle       lipgloss.Style
	WarningStyle         lipgloss.Style
	WarningCharStyle     lipgloss.Style
	InformationStyle     lipgloss.Style
	InformationCharStyle lipgloss.Style
	HintStyle            lipgloss.Style
	HintCharStyle        lipgloss.Style
}

type DocumentationStyles struct {
	Style        lipgloss.Style
	MessageStyle lipgloss.Style
}

type AutocompleteStyles struct {
	Style lipgloss.Style

	ItemStyle         lipgloss.Style
	SelectedItemStyle lipgloss.Style
}

type OverlayStyles struct {
	Styles overlay.Styles

	NotificationStyle lipgloss.Style
	RunOverlayStyle   lipgloss.Style
}
