package config

import (
	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/internal/bubbles/button"
	"go.gopad.dev/gopad/internal/bubbles/filepicker"
	"go.gopad.dev/gopad/internal/bubbles/help"
	"go.gopad.dev/gopad/internal/bubbles/list"
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
	Debug      key.Binding
}

func (k KeyMap) ButtonKeyMap() button.KeyMap {
	return button.KeyMap{
		OK: k.OK,
	}
}

func (k KeyMap) HelpView() []help.KeyMapCategory {
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
				k.Debug,
			},
		},
	}
	binds = append(binds, k.Editor.HelpView()...)
	binds = append(binds, k.FilePicker.HelpView()...)
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
	ToggleFileTree key.Binding
	FocusFileTree  key.Binding
	Search         key.Binding
	OpenOutline    key.Binding

	RefreshSyntaxHighlight key.Binding
	ToggleTreeSitterDebug  key.Binding
	DebugTreeSitterNodes   key.Binding

	File         EditorFileKeyMap
	Navigation   EditorNavigationKeyMap
	Selection    EditorSelectionKeyMap
	Edit         EditorEditKeyMap
	Code         EditorCodeKeyMap
	Autocomplete EditorAutocompleteKeyMap
	Diagnostic   EditorDiagnosticKeyMap

	FileTree  FileTreeKeyMap
	SearchBar SearchbarKeyMap
}

func (k EditorKeyMap) HelpView() []help.KeyMapCategory {
	return []help.KeyMapCategory{
		{
			Category: "Editor",
			Keys: []key.Binding{
				k.ToggleFileTree,
				k.FocusFileTree,
				k.Search,
				k.OpenOutline,
				emptyKeyBind,
				k.RefreshSyntaxHighlight,
				k.ToggleTreeSitterDebug,
				k.DebugTreeSitterNodes,
			},
		},
		k.File.HelpView(),
		k.Navigation.HelpView(),
		k.Selection.HelpView(),
		k.Edit.HelpView(),
		k.Code.HelpView(),
		k.Autocomplete.HelpView(),
		k.Diagnostic.HelpView(),
		k.FileTree.HelpView(),
		k.SearchBar.HelpView(),
	}
}

func (k EditorKeyMap) TextInputKeyMap() textinput.KeyMap {
	return textinput.KeyMap{
		CharacterLeft:  k.Navigation.CharacterLeft,
		CharacterRight: k.Navigation.CharacterRight,
		WordLeft:       k.Navigation.WordLeft,
		WordRight:      k.Navigation.WordRight,
		LineStart:      k.Navigation.LineStart,
		LineEnd:        k.Navigation.LineEnd,
		DeleteLeft:     k.Edit.DeleteLeft,
		DeleteRight:    k.Edit.DeleteRight,
		Paste:          k.Edit.Paste,
	}
}

type EditorFileKeyMap struct {
	Open       key.Binding
	OpenFolder key.Binding
	Close      key.Binding

	New    key.Binding
	Rename key.Binding
	Save   key.Binding
	Delete key.Binding

	Next key.Binding
	Prev key.Binding
}

func (k EditorFileKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor File",
		Keys: []key.Binding{
			k.Open,
			k.OpenFolder,
			k.Close,
			emptyKeyBind,
			k.New,
			k.Rename,
			k.Save,
			k.Delete,
			emptyKeyBind,
			k.Next,
			k.Prev,
		},
	}
}

type EditorNavigationKeyMap struct {
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

	GoTo key.Binding
}

func (k EditorNavigationKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor Navigation",
		Keys: []key.Binding{
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
			k.GoTo,
		},
	}
}

type EditorSelectionKeyMap struct {
	SelectLeft  key.Binding
	SelectRight key.Binding
	SelectUp    key.Binding
	SelectDown  key.Binding

	SelectAll key.Binding
}

func (k EditorSelectionKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor Selection",
		Keys: []key.Binding{
			k.SelectLeft,
			k.SelectRight,
			k.SelectUp,
			k.SelectDown,
			emptyKeyBind,
			k.SelectAll,
		},
	}
}

type EditorEditKeyMap struct {
	Tab       key.Binding
	RemoveTab key.Binding

	Paste key.Binding
	Copy  key.Binding
	Cut   key.Binding

	Undo key.Binding
	Redo key.Binding

	DeleteLeft      key.Binding
	DeleteRight     key.Binding
	DeleteWordLeft  key.Binding
	DeleteWordRight key.Binding

	DuplicateLine key.Binding
	Newline       key.Binding
	DeleteLine    key.Binding

	ToggleComment key.Binding
}

func (k EditorEditKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor Edit",
		Keys: []key.Binding{
			k.Tab,
			k.RemoveTab,
			emptyKeyBind,
			k.Cut,
			k.Copy,
			k.Paste,
			emptyKeyBind,
			k.Undo,
			k.Redo,
			emptyKeyBind,
			k.DeleteLeft,
			k.DeleteRight,
			k.DeleteWordLeft,
			k.DeleteWordRight,
			emptyKeyBind,
			k.DuplicateLine,
			k.DeleteLine,
			emptyKeyBind,
			k.ToggleComment,
		},
	}
}

type EditorCodeKeyMap struct {
	ShowDeclaration    key.Binding
	ShowDefinitions    key.Binding
	ShowTypeDefinition key.Binding
	ShowImplementation key.Binding
	ShowReferences     key.Binding
}

func (k EditorCodeKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor Code",
		Keys: []key.Binding{
			k.ShowDeclaration,
			k.ShowDefinitions,
			k.ShowTypeDefinition,
			k.ShowImplementation,
			k.ShowReferences,
		},
	}
}

type EditorAutocompleteKeyMap struct {
	Show  key.Binding
	Next  key.Binding
	Prev  key.Binding
	Apply key.Binding
}

func (k EditorAutocompleteKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor Show",
		Keys: []key.Binding{
			k.Show,
			k.Next,
			k.Prev,
			k.Apply,
		},
	}
}

type EditorDiagnosticKeyMap struct {
	Show key.Binding
	Next key.Binding
	Prev key.Binding
}

func (k EditorDiagnosticKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor Diagnostic",
		Keys: []key.Binding{
			k.Show,
			k.Next,
			k.Prev,
		},
	}
}

type FileTreeKeyMap struct {
	SelectPrev  key.Binding
	SelectNext  key.Binding
	ExpandWidth key.Binding
	ShrinkWidth key.Binding
	Open        key.Binding
	Refresh     key.Binding
}

func (k FileTreeKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Editor File Tree",
		Keys: []key.Binding{
			k.SelectPrev,
			k.SelectNext,
			emptyKeyBind,
			k.ExpandWidth,
			k.ShrinkWidth,
			emptyKeyBind,
			k.Open,
			k.Refresh,
		},
	}
}

type SearchbarKeyMap struct {
	SelectPrev key.Binding
	SelectNext key.Binding

	SelectResult key.Binding
	Close        key.Binding
}

func (k SearchbarKeyMap) HelpView() help.KeyMapCategory {
	return help.KeyMapCategory{
		Category: "Searchbar",
		Keys: []key.Binding{
			k.SelectPrev,
			k.SelectNext,
			emptyKeyBind,
			k.SelectResult,
			k.Close,
		},
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
	Debug      string              `toml:"debug"`
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
		Debug: key.NewBinding(
			key.WithKeys(k.Debug),
			key.WithHelp(k.Debug, "debug"),
		),
	}
}

type EditorKeyConfig struct {
	ToggleFileTree string `toml:"toggle_file_tree"`
	FocusFileTree  string `toml:"focus_file_tree"`
	Search         string `toml:"search"`
	OpenOutline    string `toml:"open_outline"`

	RefreshSyntaxHighlight string `toml:"refresh_syntax_highlight"`
	ToggleTreeSitterDebug  string `toml:"toggle_tree_sitter_debug"`
	DebugTreeSitterNodes   string `toml:"debug_tree_sitter_nodes"`

	File struct {
		Open       string `toml:"open"`
		OpenFolder string `toml:"open_folder"`
		Close      string `toml:"close"`

		New    string `toml:"new"`
		Rename string `toml:"rename"`
		Save   string `toml:"save"`
		Delete string `toml:"delete"`

		Next string `toml:"next"`
		Prev string `toml:"prev"`
	} `toml:"file"`

	Navigation struct {
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

		GoTo string `toml:"go_to"`
	} `toml:"navigation"`

	Selection struct {
		SelectLeft  string `toml:"select_left"`
		SelectRight string `toml:"select_right"`
		SelectUp    string `toml:"select_up"`
		SelectDown  string `toml:"select_down"`

		SelectAll string `toml:"select_all"`
	} `toml:"selection"`

	Edit struct {
		Tab       string `toml:"tab"`
		RemoveTab string `toml:"remove_tab"`

		Paste string `toml:"paste"`
		Copy  string `toml:"copy"`
		Cut   string `toml:"cut"`

		Undo string `toml:"undo"`
		Redo string `toml:"redo"`

		DeleteLeft      string `toml:"delete_left"`
		DeleteRight     string `toml:"delete_right"`
		DeleteWordLeft  string `toml:"delete_word_left"`
		DeleteWordRight string `toml:"delete_word_right"`

		DuplicateLine string `toml:"duplicate_line"`
		DeleteLine    string `toml:"delete_line"`

		ToggleComment string `toml:"toggle_comment"`
	} `toml:"edit"`

	Code struct {
		ShowDeclaration    string `toml:"show_declaration"`
		ShowDefinitions    string `toml:"show_definitions"`
		ShowTypeDefinition string `toml:"show_type_definition"`
		ShowImplementation string `toml:"show_implementation"`
		ShowReferences     string `toml:"show_references"`
	} `toml:"code"`

	Autocomplete struct {
		Show  string `toml:"show"`
		Next  string `toml:"next"`
		Prev  string `toml:"prev"`
		Apply string `toml:"apply"`
	} `toml:"autocomplete"`

	Diagnostic struct {
		Show string `toml:"show"`
		Next string `toml:"next"`
		Prev string `toml:"prev"`
	} `toml:"diagnostic"`

	FileTree  FileTreeKeyConfig  `toml:"file_tree"`
	SearchBar SearchBarKeyConfig `toml:"search_bar"`
}

func (k EditorKeyConfig) KeyMap() EditorKeyMap {
	return EditorKeyMap{
		ToggleFileTree: key.NewBinding(
			key.WithKeys(k.ToggleFileTree),
			key.WithHelp(k.ToggleFileTree, "toggle file tree"),
		),
		FocusFileTree: key.NewBinding(
			key.WithKeys(k.FocusFileTree),
			key.WithHelp(k.FocusFileTree, "focus file tree"),
		),
		Search: key.NewBinding(
			key.WithKeys(k.Search),
			key.WithHelp(k.Search, "search in file"),
		),
		OpenOutline: key.NewBinding(
			key.WithKeys(k.OpenOutline),
			key.WithHelp(k.OpenOutline, "open outline"),
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

		File: EditorFileKeyMap{
			Open: key.NewBinding(
				key.WithKeys(k.File.Open),
				key.WithHelp(k.File.Open, "open file"),
			),
			OpenFolder: key.NewBinding(
				key.WithKeys(k.File.OpenFolder),
				key.WithHelp(k.File.OpenFolder, "open folder"),
			),
			Close: key.NewBinding(
				key.WithKeys(k.File.Close),
				key.WithHelp(k.File.Close, "close file"),
			),

			New: key.NewBinding(
				key.WithKeys(k.File.New),
				key.WithHelp(k.File.New, "new file"),
			),
			Rename: key.NewBinding(
				key.WithKeys(k.File.Rename),
				key.WithHelp(k.File.Rename, "rename file"),
			),
			Save: key.NewBinding(
				key.WithKeys(k.File.Save),
				key.WithHelp(k.File.Save, "save file"),
			),
			Delete: key.NewBinding(
				key.WithKeys(k.File.Delete),
				key.WithHelp(k.File.Delete, "delete file"),
			),

			Next: key.NewBinding(
				key.WithKeys(k.File.Next),
				key.WithHelp(k.File.Next, "next file"),
			),
			Prev: key.NewBinding(
				key.WithKeys(k.File.Prev),
				key.WithHelp(k.File.Prev, "prev file"),
			),
		},
		Navigation: EditorNavigationKeyMap{
			CharacterLeft: key.NewBinding(
				key.WithKeys(k.Navigation.CharacterLeft),
				key.WithHelp(k.Navigation.CharacterLeft, "character left"),
			),
			CharacterRight: key.NewBinding(
				key.WithKeys(k.Navigation.CharacterRight),
				key.WithHelp(k.Navigation.CharacterRight, "character right"),
			),
			WordLeft: key.NewBinding(
				key.WithKeys(k.Navigation.WordLeft),
				key.WithHelp(k.Navigation.WordLeft, "word left"),
			),
			WordRight: key.NewBinding(
				key.WithKeys(k.Navigation.WordRight),
				key.WithHelp(k.Navigation.WordRight, "word right"),
			),

			LineUp: key.NewBinding(
				key.WithKeys(k.Navigation.LineUp),
				key.WithHelp(k.Navigation.LineUp, "line up"),
			),
			LineDown: key.NewBinding(
				key.WithKeys(k.Navigation.LineDown),
				key.WithHelp(k.Navigation.LineDown, "line down"),
			),
			WordUp: key.NewBinding(
				key.WithKeys(k.Navigation.WordUp),
				key.WithHelp(k.Navigation.WordUp, "word up"),
			),
			WordDown: key.NewBinding(
				key.WithKeys(k.Navigation.WordDown),
				key.WithHelp(k.Navigation.WordDown, "word down"),
			),
			PageDown: key.NewBinding(
				key.WithKeys(k.Navigation.PageDown),
				key.WithHelp(k.Navigation.PageDown, "page down"),
			),
			PageUp: key.NewBinding(
				key.WithKeys(k.Navigation.PageUp),
				key.WithHelp(k.Navigation.PageUp, "page up"),
			),

			LineStart: key.NewBinding(
				key.WithKeys(k.Navigation.LineStart),
				key.WithHelp(k.Navigation.LineStart, "line start"),
			),
			LineEnd: key.NewBinding(
				key.WithKeys(k.Navigation.LineEnd),
				key.WithHelp(k.Navigation.LineEnd, "line end"),
			),
			FileStart: key.NewBinding(
				key.WithKeys(k.Navigation.FileStart),
				key.WithHelp(k.Navigation.FileStart, "file start"),
			),
			FileEnd: key.NewBinding(
				key.WithKeys(k.Navigation.FileEnd),
				key.WithHelp(k.Navigation.FileEnd, "file end"),
			),

			GoTo: key.NewBinding(
				key.WithKeys(k.Navigation.GoTo),
				key.WithHelp(k.Navigation.GoTo, "go to"),
			),
		},
		Selection: EditorSelectionKeyMap{
			SelectLeft: key.NewBinding(
				key.WithKeys(k.Selection.SelectLeft),
				key.WithHelp(k.Selection.SelectLeft, "select left"),
			),
			SelectRight: key.NewBinding(
				key.WithKeys(k.Selection.SelectRight),
				key.WithHelp(k.Selection.SelectRight, "select right"),
			),
			SelectUp: key.NewBinding(
				key.WithKeys(k.Selection.SelectUp),
				key.WithHelp(k.Selection.SelectUp, "select up"),
			),
			SelectDown: key.NewBinding(
				key.WithKeys(k.Selection.SelectDown),
				key.WithHelp(k.Selection.SelectDown, "select down"),
			),

			SelectAll: key.NewBinding(
				key.WithKeys(k.Selection.SelectAll),
				key.WithHelp(k.Selection.SelectAll, "select all"),
			),
		},
		Edit: EditorEditKeyMap{
			Tab: key.NewBinding(
				key.WithKeys(k.Edit.Tab),
				key.WithHelp(k.Edit.Tab, "add tab"),
			),
			RemoveTab: key.NewBinding(
				key.WithKeys(k.Edit.RemoveTab),
				key.WithHelp(k.Edit.RemoveTab, "remove tab"),
			),

			Paste: key.NewBinding(
				key.WithKeys(k.Edit.Paste),
				key.WithHelp(k.Edit.Paste, "paste"),
			),
			Copy: key.NewBinding(
				key.WithKeys(k.Edit.Copy),
				key.WithHelp(k.Edit.Copy, "copy"),
			),
			Cut: key.NewBinding(
				key.WithKeys(k.Edit.Cut),
				key.WithHelp(k.Edit.Cut, "cut"),
			),

			Undo: key.NewBinding(
				key.WithKeys(k.Edit.Undo),
				key.WithHelp(k.Edit.Undo, "undo"),
			),
			Redo: key.NewBinding(
				key.WithKeys(k.Edit.Redo),
				key.WithHelp(k.Edit.Redo, "redo"),
			),

			DeleteLeft: key.NewBinding(
				key.WithKeys(k.Edit.DeleteLeft),
				key.WithHelp(k.Edit.DeleteLeft, "delete before cursor"),
			),
			DeleteRight: key.NewBinding(
				key.WithKeys(k.Edit.DeleteRight),
				key.WithHelp(k.Edit.DeleteRight, "delete after cursor"),
			),
			DeleteWordLeft: key.NewBinding(
				key.WithKeys(k.Edit.DeleteLeft),
				key.WithHelp(k.Edit.DeleteLeft, "delete word before cursor"),
			),
			DeleteWordRight: key.NewBinding(
				key.WithKeys(k.Edit.DeleteRight),
				key.WithHelp(k.Edit.DeleteRight, "delete word after cursor"),
			),

			DuplicateLine: key.NewBinding(
				key.WithKeys(k.Edit.DuplicateLine),
				key.WithHelp(k.Edit.DuplicateLine, "duplicate line"),
			),
			Newline: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "newline"),
			),
			DeleteLine: key.NewBinding(
				key.WithKeys(k.Edit.DeleteLine),
				key.WithHelp(k.Edit.DeleteLine, "delete line"),
			),
			ToggleComment: key.NewBinding(
				key.WithKeys(k.Edit.ToggleComment),
				key.WithHelp(k.Edit.ToggleComment, "toggle comment"),
			),
		},
		Code: EditorCodeKeyMap{
			ShowDeclaration: key.NewBinding(
				key.WithKeys(k.Code.ShowDeclaration),
				key.WithHelp(k.Code.ShowDeclaration, "show declaration"),
			),
			ShowDefinitions: key.NewBinding(
				key.WithKeys(k.Code.ShowDefinitions),
				key.WithHelp(k.Code.ShowDefinitions, "show definitions"),
			),
			ShowTypeDefinition: key.NewBinding(
				key.WithKeys(k.Code.ShowTypeDefinition),
				key.WithHelp(k.Code.ShowTypeDefinition, "show type definition"),
			),
			ShowImplementation: key.NewBinding(
				key.WithKeys(k.Code.ShowImplementation),
				key.WithHelp(k.Code.ShowImplementation, "show implementation"),
			),
			ShowReferences: key.NewBinding(
				key.WithKeys(k.Code.ShowReferences),
				key.WithHelp(k.Code.ShowReferences, "show references"),
			),
		},
		Autocomplete: EditorAutocompleteKeyMap{
			Show: key.NewBinding(
				key.WithKeys(k.Autocomplete.Show),
				key.WithHelp(k.Autocomplete.Show, "show autocomplete"),
			),
			Next: key.NewBinding(
				key.WithKeys(k.Autocomplete.Next),
				key.WithHelp(k.Autocomplete.Next, "next completion"),
			),
			Prev: key.NewBinding(
				key.WithKeys(k.Autocomplete.Prev),
				key.WithHelp(k.Autocomplete.Prev, "prev completion"),
			),
			Apply: key.NewBinding(
				key.WithKeys(k.Autocomplete.Apply),
				key.WithHelp(k.Autocomplete.Apply, "apply completion"),
			),
		},
		Diagnostic: EditorDiagnosticKeyMap{
			Show: key.NewBinding(
				key.WithKeys(k.Diagnostic.Show),
				key.WithHelp(k.Diagnostic.Show, "show current diagnostic"),
			),
			Next: key.NewBinding(
				key.WithKeys(k.Diagnostic.Next),
				key.WithHelp(k.Diagnostic.Next, "show next diagnostic"),
			),
			Prev: key.NewBinding(
				key.WithKeys(k.Diagnostic.Prev),
				key.WithHelp(k.Diagnostic.Prev, "show prev diagnostic"),
			),
		},

		FileTree: FileTreeKeyMap{
			SelectPrev: key.NewBinding(
				key.WithKeys(k.FileTree.SelectPrev),
				key.WithHelp(k.FileTree.SelectPrev, "select prev"),
			),
			SelectNext: key.NewBinding(
				key.WithKeys(k.FileTree.SelectNext),
				key.WithHelp(k.FileTree.SelectNext, "select next"),
			),
			ExpandWidth: key.NewBinding(
				key.WithKeys(k.FileTree.ExpandWidth),
				key.WithHelp(k.FileTree.ExpandWidth, "expand width"),
			),
			ShrinkWidth: key.NewBinding(
				key.WithKeys(k.FileTree.ShrinkWidth),
				key.WithHelp(k.FileTree.ShrinkWidth, "shrink width"),
			),
			Open: key.NewBinding(
				key.WithKeys(k.FileTree.Open),
				key.WithHelp(k.FileTree.Open, "open file or directory"),
			),
			Refresh: key.NewBinding(
				key.WithKeys(k.FileTree.Refresh),
				key.WithHelp(k.FileTree.Refresh, "refresh file tree"),
			),
		},
		SearchBar: SearchbarKeyMap{
			SelectPrev: key.NewBinding(
				key.WithKeys(k.SearchBar.SelectPrev),
				key.WithHelp(k.SearchBar.SelectPrev, "select prev"),
			),
			SelectNext: key.NewBinding(
				key.WithKeys(k.SearchBar.SelectNext),
				key.WithHelp(k.SearchBar.SelectNext, "select next"),
			),
			SelectResult: key.NewBinding(
				key.WithKeys(k.SearchBar.SelectResult),
				key.WithHelp(k.SearchBar.SelectResult, "select result"),
			),
			Close: key.NewBinding(
				key.WithKeys(k.SearchBar.Close),
				key.WithHelp(k.SearchBar.Close, "close search"),
			),
		},
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

type SearchBarKeyConfig struct {
	SelectPrev   string `toml:"select_prev"`
	SelectNext   string `toml:"select_next"`
	SelectResult string `toml:"select_result"`
	Close        string `toml:"close"`
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
