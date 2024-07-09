package config

import (
	tea "github.com/charmbracelet/bubbletea"
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
	Name       string           `toml:"name"`
	Colors     Colors           `toml:"colors"`
	Icons      IconsConfig      `toml:"icons"`
	UI         UIConfig         `toml:"ui"`
	Diagnostic DiagnosticConfig `toml:"diagnostic"`
	CodeStyles CodeStylesConfig `toml:"code_styles"`
}

func (c RawThemeConfig) Title() string {
	return c.Name
}

func (c RawThemeConfig) Description() string {
	return ""
}

func (c RawThemeConfig) Theme() ThemeConfig {
	colors := c.Colors
	return ThemeConfig{
		Name:   c.Name,
		Colors: colors,
		Icons:  c.Icons.Styles(colors),
		UI: UiStyles{
			AppBar: AppBarStyles{
				Style:      c.UI.AppBar.Style.Style(colors),
				TitleStyle: c.UI.AppBar.Title.Style(colors).Padding(0, 1),
				Files: AppBarFilesStyle{
					Style:             c.UI.AppBar.Files.Style.Style(colors),
					FileStyle:         c.UI.AppBar.Files.File.Style(colors).Padding(0, 1, 0, 2),
					SelectedFileStyle: c.UI.AppBar.Files.SelectedFile.Style(colors).Padding(0, 1, 0, 2),
				},
			},
			FileTree: filetree.Styles{
				Style:                       lipgloss.NewStyle(),
				EmptyStyle:                  lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center),
				EntryPrefixStyle:            lipgloss.NewStyle().Faint(true),
				EntryStyle:                  c.UI.Menu.Entry.Style(colors),
				EntrySelectedStyle:          c.UI.Menu.SelectedEntry.Style(colors),
				EntrySelectedUnfocusedStyle: c.UI.Menu.SelectedEntryUnfocused.Style(colors),
			},
			FileView: FileViewStyles{
				Style:                  c.UI.FileView.Style.Style(colors),
				EmptyStyle:             c.UI.FileView.Empty.Style(colors).Align(lipgloss.Center, lipgloss.Center),
				BorderStyle:            c.UI.FileView.Border.Style(colors).Border(lipgloss.NormalBorder(), false, false, false, true),
				LineStyle:              c.UI.FileView.Line.Style(colors),
				LinePrefixStyle:        c.UI.FileView.LinePrefix.Style(colors).Padding(0, 1),
				LineCharStyle:          c.UI.FileView.LineChar.Style(colors),
				CurrentLineStyle:       c.UI.FileView.CurrentLine.Style(colors),
				CurrentLinePrefixStyle: c.UI.FileView.CurrentLinePrefix.Style(colors).Padding(0, 1),
				CurrentLineCharStyle:   c.UI.FileView.CurrentLineChar.Style(colors),
				SelectionStyle:         c.UI.FileView.Selection.Style(colors),
				InlayHintStyle:         c.UI.FileView.InlayHint.Style(colors),
			},
			CodeBar: CodeBarStyles{
				Style: c.UI.CodeBar.Style.Style(colors).Padding(0, 1),
			},
			SearchBar: searchbar.Styles{
				Style:       lipgloss.NewStyle().Padding(0, 2),
				ResultStyle: lipgloss.NewStyle().Padding(0, 1),
			},

			Documentation: DocumentationStyles{
				Style: c.UI.Menu.Style.Style(colors).Padding(0, 1),
			},
			Autocomplete: AutocompleteStyles{
				Style:             c.UI.Menu.Style.Style(colors).Padding(0, 1),
				ItemStyle:         c.UI.Menu.Entry.Style(colors),
				SelectedItemStyle: c.UI.Menu.SelectedEntry.Style(colors),
			},

			Overlay: OverlayStyles{
				Styles: overlay.Styles{
					Style:        c.UI.Menu.Style.Style(colors).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()),
					TitleStyle:   c.UI.Menu.Title.Style(colors).AlignHorizontal(lipgloss.Center).Margin(0, 1),
					ContentStyle: c.UI.Menu.Content.Style(colors).Padding(1, 2),
				},

				RunOverlayStyle: c.UI.Menu.Style.Style(colors).Align(lipgloss.Top, lipgloss.Center).Border(lipgloss.RoundedBorder()).Padding(0, 1),
			},

			TextInput: textinput.Styles{
				PromptStyle:        lipgloss.NewStyle(),
				FocusedPromptStyle: lipgloss.NewStyle(),
				TextStyle:          lipgloss.NewStyle(),
				PlaceholderStyle:   lipgloss.NewStyle(),
			},
			Button: button.Styles{
				Default: lipgloss.NewStyle().Padding(0, 1).Margin(0, 1),
				Focus:   lipgloss.NewStyle().Padding(0, 1).Margin(0, 1),
			},
			FilePicker: filepicker.Styles{
				DisabledCursor:   lipgloss.NewStyle(),
				Cursor:           lipgloss.NewStyle().Bold(true),
				Symlink:          lipgloss.NewStyle(),
				Directory:        lipgloss.NewStyle(),
				File:             lipgloss.NewStyle(),
				DisabledFile:     lipgloss.NewStyle(),
				DisabledSelected: lipgloss.NewStyle(),
				Permission:       lipgloss.NewStyle(),
				Selected:         lipgloss.NewStyle().Bold(true),
				FileSize:         lipgloss.NewStyle().Width(7).Align(lipgloss.Right),
				EmptyDirectory:   lipgloss.NewStyle().PaddingLeft(2),
			},
			Cursor: cursor.Styles{
				BlockCursor:     c.UI.Cursor.Block.Style(colors),
				UnderlineCursor: c.UI.Cursor.Underline.Style(colors).Underline(true),
			},
			Help: help.Styles{
				Ellipsis:       lipgloss.NewStyle(),
				Header:         lipgloss.NewStyle().AlignHorizontal(lipgloss.Center),
				ShortKey:       lipgloss.NewStyle(),
				ShortDesc:      lipgloss.NewStyle(),
				ShortSeparator: lipgloss.NewStyle(),
				FullKey:        lipgloss.NewStyle(),
				FullDesc:       lipgloss.NewStyle(),
				FullSeparator:  lipgloss.NewStyle(),
			},
			NotificationStyle: notifications.Styles{
				Notification: c.UI.Menu.Style.Style(colors).MaxWidth(32).Border(lipgloss.RoundedBorder()).Padding(0, 1),
			},
			List: list.Styles{
				Style:             c.UI.Menu.Style.Style(colors).MarginLeft(1),
				ItemStyle:         c.UI.Menu.Entry.Style(colors).Padding(0, 1),
				ItemSelectedStyle: c.UI.Menu.SelectedEntry.Style(colors).Padding(0, 1),

				ItemDescriptionStyle: lipgloss.NewStyle(),
			},
		},

		Diagnostic: DiagnosticStyles{
			ErrorStyle:           c.Diagnostic.Error.Style(colors),
			ErrorCharStyle:       c.Diagnostic.ErrorChar.Style(colors),
			WarningStyle:         c.Diagnostic.Warning.Style(colors),
			WarningCharStyle:     c.Diagnostic.WarningChar.Style(colors),
			InfoStyle:            c.Diagnostic.Info.Style(colors),
			InfoCharStyle:        c.Diagnostic.InfoChar.Style(colors),
			HintStyle:            c.Diagnostic.Hint.Style(colors),
			HintCharStyle:        c.Diagnostic.HintChar.Style(colors),
			DeprecatedStyle:      c.Diagnostic.Deprecated.Style(colors),
			DeprecatedCharStyle:  c.Diagnostic.DeprecatedChar.Style(colors),
			UnnecessaryStyle:     c.Diagnostic.Unnecessary.Style(colors),
			UnnecessaryCharStyle: c.Diagnostic.UnnecessaryChar.Style(colors),
		},
		CodeStyles: c.CodeStyles.Styles(colors),
	}
}

type Colors map[string]lipgloss.Color

type IconsConfig struct {
	RootDir IconConfig `toml:"root_dir"`
	Dir     IconConfig `toml:"dir"`
	OpenDir IconConfig `toml:"open_dir"`
	File    IconConfig `toml:"file"`

	Error   IconConfig `toml:"error"`
	Warning IconConfig `toml:"warning"`
	Info    IconConfig `toml:"info"`
	Hint    IconConfig `toml:"hint"`

	Files map[string]IconConfig `toml:"files"`

	UnknownType IconConfig            `toml:"unknown_type"`
	Types       map[string]IconConfig `toml:"types"`
}

func (c IconsConfig) Styles(colors Colors) IconStyles {
	files := make(map[string]lipgloss.Style, len(c.Files))
	for k, v := range c.Files {
		files[k] = v.IconStyle(colors)
	}

	types := make(map[string]lipgloss.Style, len(c.Types))
	for k, v := range c.Types {
		types[k] = v.IconStyle(colors)
	}

	return IconStyles{
		RootDir: c.RootDir.IconStyle(colors),
		Dir:     c.Dir.IconStyle(colors),
		OpenDir: c.OpenDir.IconStyle(colors),
		File:    c.File.IconStyle(colors),

		Error:   c.Error.IconStyle(colors),
		Warning: c.Warning.IconStyle(colors),
		Info:    c.Info.IconStyle(colors),
		Hint:    c.Hint.IconStyle(colors),

		Files: files,

		UnknownType: c.UnknownType.IconStyle(colors),
		Types:       types,
	}
}

type IconConfig struct {
	Icon  rune  `toml:"icon"`
	Style Style `toml:"style"`
}

func (c IconConfig) IconStyle(colors Colors) lipgloss.Style {
	return c.Style.Style(colors).SetString(string(c.Icon))
}

type UIConfig struct {
	AppBar  AppBarUIConfig  `toml:"app_bar"`
	CodeBar CodeBarUIConfig `toml:"code_bar"`

	Menu   MenuUIConfig   `toml:"menu"`
	Cursor CursorUIConfig `toml:"cursor"`

	FileTree FileTreeUIConfig `toml:"file_tree"`
	FileView FileViewUIConfig `toml:"file_view"`
}

type AppBarUIConfig struct {
	Style Style `toml:"style"`
	Title Style `toml:"title"`

	Files AppBarFilesUIConfig `toml:"files"`
}

type AppBarFilesUIConfig struct {
	Style        Style `toml:"style"`
	File         Style `toml:"file"`
	SelectedFile Style `toml:"selected_file"`
}

type CodeBarUIConfig struct {
	Style Style `toml:"style"`
}

type MenuUIConfig struct {
	Style   Style `toml:"style"`
	Title   Style `toml:"title"`
	Content Style `toml:"content"`

	Entry                  Style `toml:"entry"`
	SelectedEntry          Style `toml:"selected_entry"`
	SelectedEntryUnfocused Style `toml:"selected_entry_unfocused"`
}

type CursorUIConfig struct {
	Block     Style `toml:"block"`
	Underline Style `toml:"underline"`
}

type FileTreeUIConfig struct {
}

type FileViewUIConfig struct {
	Style  Style `toml:"style"`
	Empty  Style `toml:"empty"`
	Border Style `toml:"border"`

	Line       Style `toml:"line"`
	LinePrefix Style `toml:"line_prefix"`
	LineChar   Style `toml:"line_char"`

	CurrentLine       Style `toml:"current_line"`
	CurrentLinePrefix Style `toml:"current_line_prefix"`
	CurrentLineChar   Style `toml:"current_line_char"`

	Selection Style `toml:"selection"`
	InlayHint Style `toml:"inlay_hint"`
}

type DiagnosticConfig struct {
	Error     Style `toml:"error"`
	ErrorChar Style `toml:"error_char"`

	Warning     Style `toml:"warning"`
	WarningChar Style `toml:"warning_char"`

	Info     Style `toml:"info"`
	InfoChar Style `toml:"info_char"`

	Hint     Style `toml:"hint"`
	HintChar Style `toml:"hint_char"`

	Deprecated     Style `toml:"deprecated"`
	DeprecatedChar Style `toml:"deprecated_char"`

	Unnecessary     Style `toml:"unnecessary"`
	UnnecessaryChar Style `toml:"unnecessary_char"`
}

type CodeStylesConfig map[string]Style

func (c CodeStylesConfig) Styles(colors Colors) map[string]lipgloss.Style {
	m := make(map[string]lipgloss.Style, len(c))
	for k, v := range c {
		m[k] = v.Style(colors)
	}

	tea.HideCursor()

	return m
}
