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

type RawThemeConfig struct {
	Name   string       `toml:"name"`
	Colors ColorsConfig `toml:"colors"`
	Icons  IconsConfig  `toml:"icons"`
	Styles StylesConfig `toml:"styles"`
	Code   CodeStyles   `toml:"code"`
}

func (c RawThemeConfig) Title() string {
	return c.Name
}

func (c RawThemeConfig) Description() string {
	return ""
}

func (c RawThemeConfig) Theme() ThemeConfig {
	colors := c.Colors.Colors()
	return ThemeConfig{
		Name:             c.Name,
		Colors:           colors,
		Icons:            c.Icons,
		AppBarStyle:      lipgloss.NewStyle().Foreground(colors.PrimaryColor).Reverse(true),
		AppBarTitleStyle: lipgloss.NewStyle().Padding(0, 1),
		Editor: EditorStyles{
			EmptyStyle: lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center),

			FileStyle:         lipgloss.NewStyle().Foreground(colors.PrimaryColor).Padding(0, 1, 0, 2).Reverse(true),
			FileSelectedStyle: lipgloss.NewStyle().Foreground(colors.PrimarySelectedColor).Padding(0, 1, 0, 2).Reverse(true),

			CodeBorderStyle: lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(colors.PrimaryColor),

			CodeLineStyle:     lipgloss.NewStyle(),
			CodePrefixStyle:   lipgloss.NewStyle().Foreground(colors.SecondaryTextColor).Bold(true).Padding(0, 1),
			CodeLineCharStyle: lipgloss.NewStyle(),

			CodeCurrentLineStyle:       lipgloss.NewStyle().Background(colors.SecondaryBackgroundColor),
			CodeCurrentLinePrefixStyle: lipgloss.NewStyle().Foreground(colors.PrimaryTextColor).Bold(true).Background(colors.SecondaryBackgroundColor).Padding(0, 1),
			CodeCurrentLineCharStyle:   lipgloss.NewStyle().Background(colors.SecondaryBackgroundColor),

			CodeSelectionStyle: lipgloss.NewStyle().Reverse(true),
			CodeBarStyle:       lipgloss.NewStyle().Foreground(colors.PrimaryColor).Reverse(true).Padding(0, 1),
			CodeInlayHintStyle: lipgloss.NewStyle().Foreground(colors.SecondaryTextColor).Background(colors.SecondaryBackgroundColor).Bold(true).Italic(true),

			FileTree: filetree.Styles{
				Style:                       lipgloss.NewStyle(),
				EmptyStyle:                  lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center),
				EntryPrefixStyle:            lipgloss.NewStyle().Faint(true),
				EntryStyle:                  lipgloss.NewStyle(),
				EntrySelectedStyle:          lipgloss.NewStyle().Foreground(colors.PrimarySelectedColor).Reverse(true),
				EntrySelectedUnfocusedStyle: lipgloss.NewStyle().Foreground(colors.PrimaryColor).Reverse(true),
			},
			SearchBar: searchbar.Styles{
				Style:       lipgloss.NewStyle().Padding(0, 2),
				ResultStyle: lipgloss.NewStyle().Padding(0, 1),
			},
			Diagnostics: DiagnosticStyles{
				ErrorStyle:           lipgloss.NewStyle().Foreground(colors.ErrorColor).Bold(true),
				ErrorCharStyle:       c.Styles.Diagnostics.Error.Style(),
				WarningStyle:         lipgloss.NewStyle().Foreground(colors.WarningColor).Bold(true),
				WarningCharStyle:     c.Styles.Diagnostics.Warning.Style(),
				InformationStyle:     lipgloss.NewStyle().Foreground(colors.InformationColor).Bold(true),
				InformationCharStyle: c.Styles.Diagnostics.Information.Style(),
				HintStyle:            lipgloss.NewStyle().Foreground(colors.HintColor).Bold(true),
				HintCharStyle:        c.Styles.Diagnostics.Hint.Style(),
			},
			Documentation: DocumentationStyles{
				Style:        lipgloss.NewStyle().Background(colors.SecondaryBackgroundColor).Padding(0, 1),
				MessageStyle: lipgloss.NewStyle().Background(colors.SecondaryBackgroundColor),
			},
			Autocomplete: AutocompleteStyles{
				Style: lipgloss.NewStyle().Background(colors.SecondaryBackgroundColor).Padding(0, 1),

				ItemStyle:         lipgloss.NewStyle(),
				SelectedItemStyle: lipgloss.NewStyle().Reverse(true),
			},
			CodeStyles: c.Code.Styles(),
		},
		Overlay: OverlayStyles{
			Styles: overlay.Styles{
				Style:        lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).BorderForeground(colors.PrimaryColor),
				TitleStyle:   lipgloss.NewStyle().Foreground(colors.PrimaryColor).Reverse(true).AlignHorizontal(lipgloss.Center).Margin(0, 1),
				ContentStyle: lipgloss.NewStyle().Padding(1, 2),
			},

			NotificationStyle: lipgloss.NewStyle().Foreground(colors.PrimaryTextColor).Width(16).Border(lipgloss.RoundedBorder()).BorderForeground(colors.PrimaryColor).Padding(0, 1),

			RunOverlayStyle: lipgloss.NewStyle().Align(lipgloss.Top, lipgloss.Center).Border(lipgloss.RoundedBorder()).BorderForeground(colors.PrimaryColor).Padding(0, 1),
		},
		TextInput: textinput.Styles{
			PromptStyle:        lipgloss.NewStyle().Foreground(colors.SecondaryTextColor),
			FocusedPromptStyle: lipgloss.NewStyle().Foreground(colors.PrimaryTextColor),
			TextStyle:          lipgloss.NewStyle().Foreground(colors.PrimaryTextColor),
			PlaceholderStyle:   lipgloss.NewStyle().Foreground(colors.SecondaryTextColor),
		},
		Button: button.Styles{
			Default: lipgloss.NewStyle().Foreground(colors.PrimaryColor).Reverse(true).Padding(0, 1).Margin(0, 1),
			Focus:   lipgloss.NewStyle().Foreground(colors.PrimarySelectedColor).Reverse(true).Padding(0, 1).Margin(0, 1),
		},
		FilePicker: filepicker.Styles{
			DisabledCursor:   lipgloss.NewStyle().Foreground(colors.DisabledCursorColor),
			Cursor:           lipgloss.NewStyle().Foreground(colors.PrimarySelectedColor).Bold(true),
			Symlink:          lipgloss.NewStyle().Foreground(colors.InformationColor),
			Directory:        lipgloss.NewStyle().Foreground(colors.PrimaryColor),
			File:             lipgloss.NewStyle().Foreground(colors.PrimaryTextColor),
			DisabledFile:     lipgloss.NewStyle().Foreground(colors.DisabledTextColor),
			DisabledSelected: lipgloss.NewStyle().Foreground(colors.DisabledTextColor),
			Permission:       lipgloss.NewStyle().Foreground(colors.DisabledTextColor),
			Selected:         lipgloss.NewStyle().Foreground(colors.PrimarySelectedColor).Bold(true),
			FileSize:         lipgloss.NewStyle().Foreground(colors.DisabledTextColor).Width(7).Align(lipgloss.Right),
			EmptyDirectory:   lipgloss.NewStyle().Foreground(colors.ErrorColor).PaddingLeft(2).SetString("No Files Found."),
		},
		Cursor: cursor.Styles{
			BlockCursor:     lipgloss.NewStyle().Foreground(colors.CursorColor).Reverse(true),
			UnderlineCursor: lipgloss.NewStyle().Underline(true),
		},
		Help: help.Styles{
			Ellipsis:       lipgloss.NewStyle(),
			Header:         lipgloss.NewStyle().Reverse(true).AlignHorizontal(lipgloss.Center),
			ShortKey:       lipgloss.NewStyle(),
			ShortDesc:      lipgloss.NewStyle(),
			ShortSeparator: lipgloss.NewStyle(),
			FullKey:        lipgloss.NewStyle(),
			FullDesc:       lipgloss.NewStyle(),
			FullSeparator:  lipgloss.NewStyle(),
		},
		NotificationStyle: notifications.Styles{
			Notification: lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder()).BorderForeground(colors.PrimaryColor),
		},
		List: list.Styles{
			Style:             lipgloss.NewStyle().MarginLeft(1),
			ItemStyle:         lipgloss.NewStyle().Padding(0, 1),
			ItemSelectedStyle: lipgloss.NewStyle().Padding(0, 1).Reverse(true),

			ItemDescriptionStyle: lipgloss.NewStyle().Foreground(colors.SecondaryTextColor),
		},
	}
}

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

type StylesConfig struct {
	Diagnostics DiagnosticsStylesConfig `toml:"diagnostics"`
}

type DiagnosticsStylesConfig struct {
	Error       Style `toml:"error"`
	Warning     Style `toml:"warning"`
	Information Style `toml:"information"`
	Hint        Style `toml:"hint"`
}

type IconsConfig struct {
	RootDir rune `toml:"root_dir"`
	Dir     rune `toml:"dir"`
	OpenDir rune `toml:"open_dir"`
	File    rune `toml:"file"`

	Error       rune `toml:"error"`
	Warning     rune `toml:"warning"`
	Information rune `toml:"information"`
	Hint        rune `toml:"hint"`
}
