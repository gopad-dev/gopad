package config

import (
	"github.com/charmbracelet/bubbles/key"

	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/filetree"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/list"
	"go.gopad.dev/gopad/internal/bubbles/searchbar"
	"go.gopad.dev/gopad/internal/bubbles/textinput"
)

var emptyKeyBind = key.NewBinding(key.WithKeys([]string{}...))

type KeyMap struct {
	Quit   key.Binding
	Help   key.Binding
	OK     key.Binding
	Cancel key.Binding

	Left  key.Binding
	Right key.Binding
	Up    key.Binding
	Down  key.Binding
	Start key.Binding
	End   key.Binding

	Editor     EditorKeyMap
	FilePicker filepicker.KeyMap
	Run        key.Binding
	Terminal   key.Binding
	KeyMapper  key.Binding
}

func (k KeyMap) ButtonKeyMap() button.KeyMap {
	return button.KeyMap{
		OK: k.OK,
	}
}

func (k KeyMap) FullHelpView() []help.KeyMapCategory {
	binds := []help.KeyMapCategory{
		{
			Category: "General",
			Keys: []key.Binding{
				k.Quit,
				k.Help,
				k.OK,
				k.Cancel,
				emptyKeyBind,
				k.Left,
				k.Right,
				k.Up,
				k.Down,
				emptyKeyBind,
				k.Run,
				k.Terminal,
				k.KeyMapper,
			},
		},
	}
	binds = append(binds, k.Editor.FullHelpView()...)
	binds = append(binds, k.FilePicker.FullHelpView()...)
	return binds
}

func (k KeyMap) List() list.KeyMap {
	return list.KeyMap{
		Up:    k.Up,
		Down:  k.Down,
		Start: k.Start,
		End:   k.End,
	}
}

type EditorKeyMap struct {
	OpenFile   key.Binding
	OpenFolder key.Binding
	SaveFile   key.Binding
	CloseFile  key.Binding
	NewFile    key.Binding
	RenameFile key.Binding
	DeleteFile key.Binding

	Search         key.Binding
	OpenOutline    key.Binding
	ToggleFileTree key.Binding
	FocusFileTree  key.Binding
	NextFile       key.Binding
	PrevFile       key.Binding

	Autocomplete    key.Binding
	NextCompletion  key.Binding
	PrevCompletion  key.Binding
	ApplyCompletion key.Binding

	RefreshSyntaxHighlight key.Binding
	ToggleTreeSitterDebug  key.Binding
	DebugTreeSitterNodes   key.Binding
	ShowCurrentDiagnostic  key.Binding
	ShowDefinitions        key.Binding

	CharacterLeft  key.Binding
	CharacterRight key.Binding
	WordLeft       key.Binding
	WordRight      key.Binding

	LineUp   key.Binding
	LineDown key.Binding
	WordUp   key.Binding
	WordDown key.Binding
	PageUp   key.Binding
	PageDown key.Binding

	LineStart key.Binding
	LineEnd   key.Binding
	FileStart key.Binding
	FileEnd   key.Binding

	SelectLeft  key.Binding
	SelectRight key.Binding
	SelectUp    key.Binding
	SelectDown  key.Binding
	SelectAll   key.Binding

	Paste key.Binding
	Copy  key.Binding
	Cut   key.Binding

	Undo key.Binding
	Redo key.Binding

	Tab             key.Binding
	RemoveTab       key.Binding
	Newline         key.Binding
	DeleteLeft      key.Binding
	DeleteRight     key.Binding
	DeleteWordLeft  key.Binding
	DeleteWordRight key.Binding
	DuplicateLine   key.Binding
	DeleteLine      key.Binding

	ToggleComment key.Binding
	Debug         key.Binding

	FileTree  filetree.KeyMap
	SearchBar searchbar.KeyMap
}

func (k EditorKeyMap) FullHelpView() []help.KeyMapCategory {
	binds := []help.KeyMapCategory{
		{
			Category: "Editor",
			Keys: []key.Binding{
				k.OpenFile,
				k.OpenFolder,
				k.SaveFile,
				k.CloseFile,
				k.NewFile,
				k.RenameFile,
				k.DeleteFile,
				emptyKeyBind,
				k.Search,
				k.ToggleFileTree,
				k.NextFile,
				k.PrevFile,
				emptyKeyBind,
				k.Autocomplete,
				k.NextCompletion,
				k.PrevCompletion,
				k.ApplyCompletion,
				emptyKeyBind,
				k.RefreshSyntaxHighlight,
				k.ToggleTreeSitterDebug,
				k.DebugTreeSitterNodes,
				emptyKeyBind,
				k.CharacterLeft,
				k.CharacterRight,
				k.WordLeft,
				k.WordRight,
				emptyKeyBind,
				k.LineUp,
				k.LineDown,
				k.WordUp,
				k.WordDown,
				k.PageUp,
				k.PageDown,
				emptyKeyBind,
				k.LineStart,
				k.LineEnd,
				k.FileStart,
				k.FileEnd,
				emptyKeyBind,
				k.SelectLeft,
				k.SelectRight,
				k.SelectUp,
				k.SelectDown,
				k.SelectAll,
				emptyKeyBind,
				k.Cut,
				k.Copy,
				k.Paste,
				emptyKeyBind,
				k.Undo,
				k.Redo,
				emptyKeyBind,
				k.Tab,
				k.RemoveTab,
				k.Newline,
				k.DeleteLeft,
				k.DeleteRight,
				k.DeleteWordLeft,
				k.DeleteWordRight,
				k.DuplicateLine,
				k.DeleteLine,
				emptyKeyBind,
				k.ToggleComment,
				k.Debug,
			},
		},
	}
	binds = append(binds, k.FileTree.FullHelpView()...)
	binds = append(binds, k.SearchBar.FullHelpView()...)
	return binds
}

func (k EditorKeyMap) TextInputKeyMap() textinput.KeyMap {
	return textinput.KeyMap{
		CharacterLeft:  k.CharacterLeft,
		CharacterRight: k.CharacterRight,
		WordLeft:       k.WordLeft,
		WordRight:      k.WordRight,
		DeleteLeft:     k.DeleteLeft,
		DeleteRight:    k.DeleteRight,
		LineStart:      k.LineStart,
		LineEnd:        k.LineEnd,
		Paste:          k.Paste,
	}
}

func DefaultKeyMapConfig() KeyMapConfig {
	return KeyMapConfig{
		Quit:   "ctrl+q",
		Help:   "ctrl+h",
		OK:     "enter",
		Cancel: "esc",
		Left:   "left",
		Right:  "right",
		Up:     "up",
		Down:   "down",
		Start:  "home",
		End:    "end",
		Editor: EditorKeyConfig{
			OpenFile:   "ctrl+o",
			OpenFolder: "alt+o",
			SaveFile:   "ctrl+s",
			CloseFile:  "ctrl+w",
			NewFile:    "ctrl+n",
			RenameFile: "ctrl+r",
			DeleteFile: "ctrl+g",

			Search:         "ctrl+f",
			OpenOutline:    "alt+7",
			ToggleFileTree: "ctrl+b",
			FocusFileTree:  "alt+b",
			NextFile:       "alt+right",
			PrevFile:       "alt+left",

			Autocomplete:    "ctrl+@",
			NextCompletion:  "down",
			PrevCompletion:  "up",
			ApplyCompletion: "enter",

			RefreshSyntaxHighlight: "f1",
			ToggleTreeSitterDebug:  "f2",
			DebugTreeSitterNodes:   "f3",
			ShowCurrentDiagnostic:  "ctrl+j",
			ShowDefinitions:        "alt+.",

			CharacterLeft:  "left",
			CharacterRight: "right",
			WordLeft:       "ctrl+left",
			WordRight:      "ctrl+right",

			LineUp:   "up",
			LineDown: "down",
			WordUp:   "ctrl+up",
			WordDown: "ctrl+down",
			PageUp:   "pgup",
			PageDown: "pgdown",

			LineStart: "home",
			LineEnd:   "end",
			FileStart: "ctrl+home",
			FileEnd:   "ctrl+end",

			SelectLeft:  "shift+left",
			SelectRight: "shift+right",
			SelectUp:    "shift+up",
			SelectDown:  "shift+down",
			SelectAll:   "ctrl+a",

			Cut:   "ctrl+x",
			Copy:  "ctrl+c",
			Paste: "ctrl+v",

			Undo: "ctrl+z",
			Redo: "ctrl+y",

			Tab:             "tab",
			RemoveTab:       "shift+tab",
			Newline:         "enter",
			DeleteLeft:      "backspace",
			DeleteRight:     "delete",
			DeleteWordLeft:  "ctrl+backspace",
			DeleteWordRight: "ctrl+delete",
			DuplicateLine:   "ctrl+d",
			DeleteLine:      "alt+backspace",

			ToggleComment: "ctrl+_",
			Debug:         "f12",

			FileTree: FileTreeKeyConfig{
				SelectPrev:  "up",
				SelectNext:  "down",
				ExpandWidth: "ctrl+right",
				ShrinkWidth: "ctrl+left",
				Open:        "enter",
				Refresh:     "ctrl+r",
			},
			SearchBar: SearchBarKeyConfig{
				SelectPrev:   "up",
				SelectNext:   "down",
				SelectResult: "enter",
				Close:        "esc",
			},
		},
		FilePicker: FilePickerKeyConfig{
			GoToTop:  "home",
			GoToEnd:  "end",
			Up:       "up",
			Down:     "down",
			PageUp:   "pgup",
			PageDown: "pgdown",
			Back:     "left",
			Open:     "right",
			Select:   "enter",
		},
		Run:       "ctrl+k",
		Terminal:  "ctrl+t",
		KeyMapper: "f4",
	}
}

type KeyMapConfig struct {
	Quit   string `toml:"quit"`
	Help   string `toml:"help"`
	OK     string `toml:"ok"`
	Cancel string `toml:"cancel"`

	Left  string `toml:"left"`
	Right string `toml:"right"`
	Up    string `toml:"up"`
	Down  string `toml:"down"`
	Start string `toml:"start"`
	End   string `toml:"end"`

	Editor     EditorKeyConfig     `toml:"editor"`
	FilePicker FilePickerKeyConfig `toml:"file_picker"`
	Run        string              `toml:"run"`
	Terminal   string              `toml:"terminal"`
	KeyMapper  string              `toml:"key_mapper"`
}

func (k KeyMapConfig) Keys() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys(k.Quit),
			key.WithHelp(k.Quit, "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys(k.Help),
			key.WithHelp(k.Help, "toggle help"),
		),
		OK: key.NewBinding(
			key.WithKeys(k.OK),
			key.WithHelp(k.OK, "ok"),
		),
		Cancel: key.NewBinding(
			key.WithKeys(k.Cancel),
			key.WithHelp(k.Cancel, "cancel"),
		),
		Left: key.NewBinding(
			key.WithKeys(k.Left),
			key.WithHelp(k.Left, "left"),
		),
		Right: key.NewBinding(
			key.WithKeys(k.Right),
			key.WithHelp(k.Right, "right"),
		),
		Up: key.NewBinding(
			key.WithKeys(k.Up),
			key.WithHelp(k.Up, "up"),
		),
		Down: key.NewBinding(
			key.WithKeys(k.Down),
			key.WithHelp(k.Down, "down"),
		),
		Start: key.NewBinding(
			key.WithKeys(k.Start),
			key.WithHelp(k.Start, "start"),
		),
		End: key.NewBinding(
			key.WithKeys(k.End),
			key.WithHelp(k.End, "end"),
		),
		Editor:     k.Editor.KeyMap(),
		FilePicker: k.FilePicker.KeyMap(),
		Run: key.NewBinding(
			key.WithKeys(k.Run),
			key.WithHelp(k.Run, "open run"),
		),
		Terminal: key.NewBinding(
			key.WithKeys(k.Terminal),
			key.WithHelp(k.Terminal, "open terminal"),
		),
		KeyMapper: key.NewBinding(
			key.WithKeys(k.KeyMapper),
			key.WithHelp(k.KeyMapper, "open key mapper"),
		),
	}
}

type EditorKeyConfig struct {
	OpenFile   string `toml:"open_file"`
	OpenFolder string `toml:"open_folder"`
	SaveFile   string `toml:"save_file"`
	CloseFile  string `toml:"close_file"`
	NewFile    string `toml:"new_file"`
	RenameFile string `toml:"rename_file"`
	DeleteFile string `toml:"delete_file"`

	Search         string `toml:"search"`
	OpenOutline    string `toml:"open_outline"`
	ToggleFileTree string `toml:"toggle_file_tree"`
	FocusFileTree  string `toml:"focus_file_tree"`
	NextFile       string `toml:"next_file"`
	PrevFile       string `toml:"prev_file"`

	Autocomplete    string `toml:"autocomplete"`
	NextCompletion  string `toml:"next_completion"`
	PrevCompletion  string `toml:"prev_completion"`
	ApplyCompletion string `toml:"apply_completion"`

	RefreshSyntaxHighlight string `toml:"refresh_syntax_highlight"`
	ToggleTreeSitterDebug  string `toml:"toggle_tree_sitter_debug"`
	DebugTreeSitterNodes   string `toml:"debug_tree_sitter_nodes"`
	ShowCurrentDiagnostic  string `toml:"show_current_diagnostic"`
	ShowDefinitions        string `toml:"show_definitions"`

	CharacterLeft  string `toml:"character_left"`
	CharacterRight string `toml:"character_right"`
	WordLeft       string `toml:"word_left"`
	WordRight      string `toml:"word_right"`

	LineUp   string `toml:"line_up"`
	LineDown string `toml:"line_down"`
	WordUp   string `toml:"word_up"`
	WordDown string `toml:"word_down"`
	PageUp   string `toml:"page_up"`
	PageDown string `toml:"page_down"`

	LineStart string `toml:"line_start"`
	LineEnd   string `toml:"line_end"`
	FileStart string `toml:"file_start"`
	FileEnd   string `toml:"file_end"`

	SelectLeft  string `toml:"select_left"`
	SelectRight string `toml:"select_right"`
	SelectUp    string `toml:"select_up"`
	SelectDown  string `toml:"select_down"`
	SelectAll   string `toml:"select_all"`

	Cut   string `toml:"cut"`
	Copy  string `toml:"copy"`
	Paste string `toml:"paste"`

	Undo string `toml:"undo"`
	Redo string `toml:"redo"`

	Tab             string `toml:"tab"`
	RemoveTab       string `toml:"remove_tab"`
	Newline         string `toml:"newline"`
	DeleteLeft      string `toml:"delete_left"`
	DeleteRight     string `toml:"delete_right"`
	DeleteWordLeft  string `toml:"delete_word_left"`
	DeleteWordRight string `toml:"delete_word_right"`
	DuplicateLine   string `toml:"duplicate_line"`
	DeleteLine      string `toml:"delete_line"`

	ToggleComment string `toml:"toggle_comment"`
	Debug         string `toml:"debug"`

	FileTree  FileTreeKeyConfig  `toml:"file_tree"`
	SearchBar SearchBarKeyConfig `toml:"search_bar"`
}

func (k EditorKeyConfig) KeyMap() EditorKeyMap {
	return EditorKeyMap{
		OpenFile: key.NewBinding(
			key.WithKeys(k.OpenFile),
			key.WithHelp(k.OpenFile, "open file"),
		),
		OpenFolder: key.NewBinding(
			key.WithKeys(k.OpenFolder),
			key.WithHelp(k.OpenFolder, "open folder"),
		),
		SaveFile: key.NewBinding(
			key.WithKeys(k.SaveFile),
			key.WithHelp(k.SaveFile, "save file"),
		),
		CloseFile: key.NewBinding(
			key.WithKeys(k.CloseFile),
			key.WithHelp(k.CloseFile, "close file"),
		),
		NewFile: key.NewBinding(
			key.WithKeys(k.NewFile),
			key.WithHelp(k.NewFile, "new file"),
		),
		RenameFile: key.NewBinding(
			key.WithKeys(k.RenameFile),
			key.WithHelp(k.RenameFile, "rename file"),
		),
		DeleteFile: key.NewBinding(
			key.WithKeys(k.DeleteFile),
			key.WithHelp(k.DeleteFile, "delete file"),
		),

		Search: key.NewBinding(
			key.WithKeys(k.Search),
			key.WithHelp(k.Search, "search in file"),
		),
		OpenOutline: key.NewBinding(
			key.WithKeys(k.OpenOutline),
			key.WithHelp(k.OpenOutline, "open outline"),
		),
		ToggleFileTree: key.NewBinding(
			key.WithKeys(k.ToggleFileTree),
			key.WithHelp(k.ToggleFileTree, "toggle file tree"),
		),
		FocusFileTree: key.NewBinding(
			key.WithKeys(k.FocusFileTree),
			key.WithHelp(k.FocusFileTree, "focus file tree"),
		),
		NextFile: key.NewBinding(
			key.WithKeys(k.NextFile),
			key.WithHelp(k.NextFile, "next file"),
		),
		PrevFile: key.NewBinding(
			key.WithKeys(k.PrevFile),
			key.WithHelp(k.PrevFile, "prev file"),
		),

		Autocomplete: key.NewBinding(
			key.WithKeys(k.Autocomplete),
			key.WithHelp(k.Autocomplete, "autocomplete"),
		),
		NextCompletion: key.NewBinding(
			key.WithKeys(k.NextCompletion),
			key.WithHelp(k.NextCompletion, "next completion"),
		),
		PrevCompletion: key.NewBinding(
			key.WithKeys(k.PrevCompletion),
			key.WithHelp(k.PrevCompletion, "prev completion"),
		),
		ApplyCompletion: key.NewBinding(
			key.WithKeys(k.ApplyCompletion),
			key.WithHelp(k.ApplyCompletion, "apply completion"),
		),

		RefreshSyntaxHighlight: key.NewBinding(
			key.WithKeys(k.RefreshSyntaxHighlight),
			key.WithHelp(k.RefreshSyntaxHighlight, "refresh syntax highlight"),
		),
		ToggleTreeSitterDebug: key.NewBinding(
			key.WithKeys(k.ToggleTreeSitterDebug),
			key.WithHelp(k.ToggleTreeSitterDebug, "toggle tree-sitter debug"),
		),
		DebugTreeSitterNodes: key.NewBinding(
			key.WithKeys(k.DebugTreeSitterNodes),
			key.WithHelp(k.DebugTreeSitterNodes, "debug tree-sitter nodes"),
		),
		ShowCurrentDiagnostic: key.NewBinding(
			key.WithKeys(k.ShowCurrentDiagnostic),
			key.WithHelp(k.ShowCurrentDiagnostic, "show current diagnostic"),
		),
		ShowDefinitions: key.NewBinding(
			key.WithKeys(k.ShowDefinitions),
			key.WithHelp(k.ShowDefinitions, "show definitions"),
		),

		CharacterLeft: key.NewBinding(
			key.WithKeys(k.CharacterLeft),
			key.WithHelp(k.CharacterLeft, "character left"),
		),
		CharacterRight: key.NewBinding(
			key.WithKeys(k.CharacterRight),
			key.WithHelp(k.CharacterRight, "character right"),
		),
		WordLeft: key.NewBinding(
			key.WithKeys(k.WordLeft),
			key.WithHelp(k.WordLeft, "word left"),
		),
		WordRight: key.NewBinding(
			key.WithKeys(k.WordRight),
			key.WithHelp(k.WordRight, "word right"),
		),

		LineUp: key.NewBinding(
			key.WithKeys(k.LineUp),
			key.WithHelp(k.LineUp, "line up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys(k.LineDown),
			key.WithHelp(k.LineDown, "line down"),
		),
		WordUp: key.NewBinding(
			key.WithKeys(k.WordUp),
			key.WithHelp(k.WordUp, "word up"),
		),
		WordDown: key.NewBinding(
			key.WithKeys(k.WordDown),
			key.WithHelp(k.WordDown, "word down"),
		),
		PageDown: key.NewBinding(
			key.WithKeys(k.PageDown),
			key.WithHelp(k.PageDown, "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys(k.PageUp),
			key.WithHelp(k.PageUp, "page up"),
		),

		LineStart: key.NewBinding(
			key.WithKeys(k.LineStart),
			key.WithHelp(k.LineStart, "line start"),
		),
		LineEnd: key.NewBinding(
			key.WithKeys(k.LineEnd),
			key.WithHelp(k.LineEnd, "line end"),
		),
		FileStart: key.NewBinding(
			key.WithKeys(k.FileStart),
			key.WithHelp(k.FileStart, "file start"),
		),
		FileEnd: key.NewBinding(
			key.WithKeys(k.FileEnd),
			key.WithHelp(k.FileEnd, "file end"),
		),

		SelectLeft: key.NewBinding(
			key.WithKeys(k.SelectLeft),
			key.WithHelp(k.SelectLeft, "select left"),
		),
		SelectRight: key.NewBinding(
			key.WithKeys(k.SelectRight),
			key.WithHelp(k.SelectRight, "select right"),
		),
		SelectUp: key.NewBinding(
			key.WithKeys(k.SelectUp),
			key.WithHelp(k.SelectUp, "select up"),
		),
		SelectDown: key.NewBinding(
			key.WithKeys(k.SelectDown),
			key.WithHelp(k.SelectDown, "select down"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys(k.SelectAll),
			key.WithHelp(k.SelectAll, "select all"),
		),

		Cut: key.NewBinding(
			key.WithKeys(k.Cut),
			key.WithHelp(k.Cut, "cut"),
		),
		Copy: key.NewBinding(
			key.WithKeys(k.Copy),
			key.WithHelp(k.Copy, "copy"),
		),
		Paste: key.NewBinding(
			key.WithKeys(k.Paste),
			key.WithHelp(k.Paste, "paste"),
		),

		Undo: key.NewBinding(
			key.WithKeys(k.Undo),
			key.WithHelp(k.Undo, "undo"),
		),
		Redo: key.NewBinding(
			key.WithKeys(k.Redo),
			key.WithHelp(k.Redo, "redo"),
		),

		Tab: key.NewBinding(
			key.WithKeys(k.Tab),
			key.WithHelp(k.Tab, "add tab"),
		),
		RemoveTab: key.NewBinding(
			key.WithKeys(k.RemoveTab),
			key.WithHelp(k.RemoveTab, "remove tab"),
		),
		Newline: key.NewBinding(
			key.WithKeys(k.Newline),
			key.WithHelp(k.Newline, "newline"),
		),
		DeleteLeft: key.NewBinding(
			key.WithKeys(k.DeleteLeft),
			key.WithHelp(k.DeleteLeft, "backspace"),
		),
		DeleteRight: key.NewBinding(
			key.WithKeys(k.DeleteRight),
			key.WithHelp(k.DeleteRight, "delete"),
		),
		DuplicateLine: key.NewBinding(
			key.WithKeys(k.DuplicateLine),
			key.WithHelp(k.DuplicateLine, "duplicate line"),
		),
		DeleteLine: key.NewBinding(
			key.WithKeys(k.DeleteLine),
			key.WithHelp(k.DeleteLine, "delete line"),
		),
		ToggleComment: key.NewBinding(
			key.WithKeys(k.ToggleComment),
			key.WithHelp(k.ToggleComment, "toggle comment"),
		),
		Debug: key.NewBinding(
			key.WithKeys(k.Debug),
			key.WithHelp(k.Debug, "debug"),
		),
		FileTree:  k.FileTree.KeyMap(),
		SearchBar: k.SearchBar.KeyMap(),
	}
}

type FileTreeKeyConfig struct {
	SelectPrev  string `toml:"select_prev"`
	SelectNext  string `toml:"select_next"`
	ExpandWidth string `toml:"expand_width"`
	ShrinkWidth string `toml:"shrink_width"`
	Open        string `toml:"open"`
	Refresh     string `toml:"refresh"`
}

func (k FileTreeKeyConfig) KeyMap() filetree.KeyMap {
	return filetree.KeyMap{
		SelectPrev: key.NewBinding(
			key.WithKeys(k.SelectPrev),
			key.WithHelp(k.SelectPrev, "select prev"),
		),
		SelectNext: key.NewBinding(
			key.WithKeys(k.SelectNext),
			key.WithHelp(k.SelectNext, "select next"),
		),
		ExpandWidth: key.NewBinding(
			key.WithKeys(k.ExpandWidth),
			key.WithHelp(k.ExpandWidth, "expand width"),
		),
		ShrinkWidth: key.NewBinding(
			key.WithKeys(k.ShrinkWidth),
			key.WithHelp(k.ShrinkWidth, "shrink width"),
		),
		Open: key.NewBinding(
			key.WithKeys(k.Open),
			key.WithHelp(k.Open, "open file or directory"),
		),
		Refresh: key.NewBinding(
			key.WithKeys(k.Refresh),
			key.WithHelp(k.Refresh, "refresh file tree"),
		),
	}
}

type SearchBarKeyConfig struct {
	SelectPrev   string `toml:"select_prev"`
	SelectNext   string `toml:"select_next"`
	SelectResult string `toml:"select_result"`
	Close        string `toml:"close"`
}

func (k SearchBarKeyConfig) KeyMap() searchbar.KeyMap {
	return searchbar.KeyMap{
		SelectPrev: key.NewBinding(
			key.WithKeys(k.SelectPrev),
			key.WithHelp(k.SelectPrev, "select prev"),
		),
		SelectNext: key.NewBinding(
			key.WithKeys(k.SelectNext),
			key.WithHelp(k.SelectNext, "select next"),
		),
		SelectResult: key.NewBinding(
			key.WithKeys(k.SelectResult),
			key.WithHelp(k.SelectResult, "select result"),
		),
		Close: key.NewBinding(
			key.WithKeys(k.Close),
			key.WithHelp(k.Close, "close search"),
		),
	}
}

type FilePickerKeyConfig struct {
	GoToTop  string `toml:"go_to_top"`
	GoToEnd  string `toml:"go_to_end"`
	Up       string `toml:"up"`
	Down     string `toml:"down"`
	PageUp   string `toml:"page_up"`
	PageDown string `toml:"page_down"`
	Back     string `toml:"back"`
	Open     string `toml:"open"`
	Select   string `toml:"select"`
}

func (k FilePickerKeyConfig) KeyMap() filepicker.KeyMap {
	return filepicker.KeyMap{
		GoToTop: key.NewBinding(
			key.WithKeys(k.GoToTop),
			key.WithHelp(k.GoToTop, "go to top"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys(k.GoToEnd),
			key.WithHelp(k.GoToEnd, "go to end"),
		),
		Up: key.NewBinding(
			key.WithKeys(k.Up),
			key.WithHelp(k.Up, "up"),
		),
		Down: key.NewBinding(
			key.WithKeys(k.Down),
			key.WithHelp(k.Down, "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys(k.PageUp),
			key.WithHelp(k.PageUp, "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys(k.PageDown),
			key.WithHelp(k.PageDown, "page down"),
		),
		Back: key.NewBinding(
			key.WithKeys(k.Back),
			key.WithHelp(k.Back, "back"),
		),
		Open: key.NewBinding(
			key.WithKeys(k.Open),
			key.WithHelp(k.Open, "open"),
		),
		Select: key.NewBinding(
			key.WithKeys(k.Select),
			key.WithHelp(k.Select, "select"),
		),
	}
}
