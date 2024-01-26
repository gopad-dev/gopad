package config

import (
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/filetree"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/searchbar"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

type ColorsConfig struct {
	PrimaryColor         string `toml:"primary_color"`
	PrimarySelectedColor string `toml:"primary_selected_color"`

	PrimaryTextColor   string `toml:"primary_text_color"`
	SecondaryTextColor string `toml:"secondary_text_color"`
	DisabledTextColor  string `toml:"disabled_text_color"`

	BackgroundColor          string `toml:"background_color"`
	SecondaryBackgroundColor string `toml:"secondary_background_color"`

	ErrorColor       string `toml:"error_color"`
	WarningColor     string `toml:"warning_color"`
	InformationColor string `toml:"information_color"`
	HintColor        string `toml:"hint_color"`

	CursorColor         string `toml:"cursor_color"`
	DisabledCursorColor string `toml:"disabled_cursor_color"`
}

func (c ColorsConfig) Colors() Colors {
	return Colors{
		PrimaryColor:         lipgloss.Color(c.PrimaryColor),
		PrimarySelectedColor: lipgloss.Color(c.PrimarySelectedColor),

		PrimaryTextColor:   lipgloss.Color(c.PrimaryTextColor),
		SecondaryTextColor: lipgloss.Color(c.SecondaryTextColor),
		DisabledTextColor:  lipgloss.Color(c.DisabledTextColor),

		BackgroundColor:          lipgloss.Color(c.BackgroundColor),
		SecondaryBackgroundColor: lipgloss.Color(c.SecondaryBackgroundColor),

		ErrorColor:       lipgloss.Color(c.ErrorColor),
		WarningColor:     lipgloss.Color(c.WarningColor),
		InformationColor: lipgloss.Color(c.InformationColor),
		HintColor:        lipgloss.Color(c.HintColor),

		CursorColor:         lipgloss.Color(c.CursorColor),
		DisabledCursorColor: lipgloss.Color(c.DisabledCursorColor),
	}
}

type CodeStyles map[string]Style

func (c CodeStyles) Styles() map[string]lipgloss.Style {
	m := make(map[string]lipgloss.Style)
	for k, v := range c {
		m[k] = v.Style()
	}
	return m
}

type Styles struct {
	Name string

	Colors Colors

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

	ErrorColor       lipgloss.Color
	WarningColor     lipgloss.Color
	InformationColor lipgloss.Color
	HintColor        lipgloss.Color

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

	FileTree     filetree.Styles
	SearchBar    searchbar.Styles
	List         list.Styles
	Diagnostics  DiagnosticStyles
	Autocomplete AutocompleteStyles

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
