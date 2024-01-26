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

type ThemeConfig struct {
	Name   string       `toml:"name"`
	Colors ColorsConfig `toml:"colors"`
	Code   CodeStyles   `toml:"code"`
}

func (c ThemeConfig) Title() string {
	return c.Name
}

func (c ThemeConfig) Description() string {
	return ""
}

func (c ThemeConfig) Styles() Styles {
	colors := c.Colors.Colors()
	return Styles{
		Name:             c.Name,
		Colors:           colors,
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

			FileTree: filetree.Styles{
				Style:                       lipgloss.NewStyle(),
				EmptyStyle:                  lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center),
				EntryPrefixStyle:            lipgloss.NewStyle().Faint(true),
				EntryStyle:                  lipgloss.NewStyle(),
				EntrySelectedStyle:          lipgloss.NewStyle().Foreground(colors.PrimaryTextColor).Reverse(true),
				EntrySelectedUnfocusedStyle: lipgloss.NewStyle().Foreground(colors.SecondaryTextColor).Reverse(true),
			},
			SearchBar: searchbar.Styles{
				Style:       lipgloss.NewStyle().Padding(0, 2),
				ResultStyle: lipgloss.NewStyle().Padding(0, 1),
			},
			Diagnostics: DiagnosticStyles{
				ErrorStyle:           lipgloss.NewStyle().Foreground(colors.ErrorColor).Bold(true),
				ErrorCharStyle:       lipgloss.NewStyle().UnderlineSpaces(false).Underline(true).UnderlineStyle(lipgloss.UnderlineStyleCurly).UnderlineColor(colors.ErrorColor),
				WarningStyle:         lipgloss.NewStyle().Foreground(colors.WarningColor).Bold(true),
				WarningCharStyle:     lipgloss.NewStyle().UnderlineSpaces(false).Underline(true).UnderlineStyle(lipgloss.UnderlineStyleCurly).UnderlineColor(colors.WarningColor),
				InformationStyle:     lipgloss.NewStyle().Foreground(colors.InformationColor).Bold(true),
				InformationCharStyle: lipgloss.NewStyle().UnderlineSpaces(false).Underline(true).UnderlineStyle(lipgloss.UnderlineStyleCurly).UnderlineColor(colors.InformationColor),
				HintStyle:            lipgloss.NewStyle().Foreground(colors.HintColor).Bold(true),
				HintCharStyle:        lipgloss.NewStyle().UnderlineSpaces(false).Underline(true).UnderlineStyle(lipgloss.UnderlineStyleCurly).UnderlineColor(colors.HintColor),
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
