package editor

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lrstanley/bubblezone"
	"github.com/mattn/go-runewidth"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/filetree"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/searchbar"
)

const (
	moveSize = 1
	pageSize = 10

	ZoneFileLanguage   = "file.language"
	ZoneFileLineEnding = "file.lineEnding"
	ZoneFileEncoding   = "file.encoding"
	ZoneFilePrefix     = "file:"
)

var fileIconByFileNameFunc = func(name string) lipgloss.Style {
	language := file.GetLanguageByFilename(name)
	var languageName string
	if language != nil {
		languageName = language.Name
	}
	return config.Theme.Icons.FileIcon(languageName)
}

func NewEditor(workspace string, args []string) (*Editor, error) {
	var cmds []tea.Cmd

	editor := Editor{
		searchBar: config.NewSearchBar(
			func(result searchbar.Result) tea.Cmd {
				return file.Scroll(result.RowStart, result.ColStart)
			},
			file.FocusFile(""),
		),
		fileTree:  config.NewFileTree(file.OpenFile, fileIconByFileNameFunc),
		workspace: workspace,
	}

	if workspace != "" {
		if err := editor.fileTree.Open(workspace); err != nil {
			return nil, fmt.Errorf("failed to init file tree: %w", err)
		}
		editor.fileTree.Show()
		cmds = append(cmds, ls.WorkspaceOpened(workspace))
	}

	for _, arg := range args {
		stat, err := os.Stat(arg)
		if errors.Is(err, os.ErrNotExist) {
			cmd, err := editor.CreateFile(arg)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, cmd)
			continue
		}
		if err != nil {
			return nil, err
		}

		if stat.IsDir() {
			continue
		}
		cmd, err := editor.OpenFile(arg)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}

	if f := editor.File(); f != nil {
		cmds = append(cmds, f.Focus())
	} else {
		editor.fileTree.Focus()
	}

	editor.init = cmds

	return &editor, nil
}

type Editor struct {
	init             []tea.Cmd
	fileTree         filetree.Model
	workspace        string
	searchBar        searchbar.Model
	files            []*file.File
	activeFile       int
	activeFileOffset int
	focus            bool
	treeSitterDebug  bool
}

func (e Editor) Workspace() string {
	return e.workspace
}

func (e *Editor) Focused() bool {
	return e.focus
}

func (e *Editor) Focus() tea.Cmd {
	e.focus = true

	if len(e.files) == 0 {
		return nil
	}

	return e.files[e.activeFile].Focus()
}

func (e *Editor) Blur() {
	e.focus = false

	e.fileTree.Blur()
	e.searchBar.Blur()

	if len(e.files) == 0 {
		return
	}
	e.files[e.activeFile].Blur()
}

func (e *Editor) CreateFile(name string) (tea.Cmd, error) {
	if !filepath.IsAbs(name) {
		name = filepath.Join(e.workspace, name)
		name, _ = filepath.Abs(name)
	}
	if slices.ContainsFunc(e.files, func(b *file.File) bool {
		return b.Name() == name
	}) {
		return nil, nil
	}

	buff, err := buffer.New(name, bytes.NewReader(nil), "utf-8", buffer.LineEndingLF, false)
	if err != nil {
		return nil, err
	}
	f := file.NewFileWithBuffer(buff, file.ModeWrite)

	e.files = append(e.files, f)

	cmds := []tea.Cmd{
		tea.Sequence(
			ls.FileCreated(f.Name(), f.Buffer().Bytes()),
			ls.FileOpened(f.Name(), f.Buffer().Version(), f.Buffer().Bytes()),
		),
	}

	if err = f.InitTree(); err != nil {
		cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree sitter tree: %s", err.Error())))
	}
	cmds = append(cmds, ls.GetInlayHint(f.Name(), f.Range()))

	return tea.Batch(cmds...), nil
}

func (e *Editor) OpenFile(name string) (tea.Cmd, error) {
	if slices.ContainsFunc(e.files, func(b *file.File) bool {
		return b.Name() == name
	}) {
		return nil, nil
	}

	f, err := file.NewFileFromName(name)
	if err != nil {
		return nil, err
	}
	e.files = append(e.files, f)

	cmds := []tea.Cmd{
		ls.FileOpened(f.Name(), f.Buffer().Version(), f.Buffer().Bytes()),
	}

	if err = f.InitTree(); err != nil {
		cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree sitter tree: %s", err.Error())))
	}
	cmds = append(cmds, ls.GetInlayHint(f.Name(), f.Range()))

	return tea.Batch(cmds...), nil
}

func (e *Editor) SaveFile(name string) (tea.Cmd, error) {
	f := e.FileByName(name)
	if f == nil {
		return nil, nil
	}
	if err := f.Buffer().Save(); err != nil {
		return nil, err
	}

	return ls.FileSaved(f.Name(), f.Buffer().Bytes()), nil
}

func (e *Editor) RenameFile(oldName string, newName string) (tea.Cmd, error) {
	if !filepath.IsAbs(newName) {
		newName = filepath.Join(e.workspace, newName)
		newName, _ = filepath.Abs(newName)
	}
	f := e.FileByName(oldName)
	if f == nil {
		return nil, nil
	}
	if err := f.Buffer().Rename(newName); err != nil {
		return nil, err
	}

	return ls.FileRenamed(f.Name(), newName), nil
}

func (e *Editor) CloseFile(name string) (tea.Cmd, error) {
	index := slices.IndexFunc(e.files, func(file *file.File) bool {
		return file.Name() == name
	})
	if index == -1 {
		return nil, nil
	}

	f := e.files[index]
	e.files = slices.Delete(e.files, index, index+1)
	e.activeFile = min(e.activeFile, len(e.files)-1)
	if len(e.files) > 0 {
		e.files[e.activeFile].Focus()
	} else {
		e.fileTree.Focus()
	}

	return ls.FileClosed(f.Name()), nil
}

func (e *Editor) DeleteFile(name string) (tea.Cmd, error) {
	index := slices.IndexFunc(e.files, func(file *file.File) bool {
		return file.Name() == name
	})
	if index == -1 {
		return nil, nil
	}

	f := e.files[index]
	if err := f.Buffer().Delete(); err != nil {
		return nil, err
	}

	e.files = slices.Delete(e.files, e.activeFile, e.activeFile+1)
	e.activeFile = min(e.activeFile, len(e.files)-1)
	if len(e.files) > 0 {
		e.files[e.activeFile].Focus()
	} else {
		e.fileTree.Focus()
	}

	return ls.FileDeleted(f.Name()), nil
}

func (e *Editor) File() *file.File {
	if len(e.files) == 0 {
		return nil
	}
	return e.files[e.activeFile]
}

func (e *Editor) SetFile(index int) {
	e.activeFile = index
}

func (e *Editor) SetFileByName(name string) {
	for i, f := range e.files {
		if f.Name() == name {
			e.activeFile = i
			return
		}
	}
}

func (e *Editor) FileByName(name string) *file.File {
	for _, f := range e.files {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

func (e *Editor) HasChanges() bool {
	for _, f := range e.files {
		if f.Dirty() {
			return true
		}
	}
	return false
}

func (e *Editor) ToggleTreeSitterDebug() {
	e.treeSitterDebug = !e.treeSitterDebug
}

func (e Editor) Init(ctx tea.Context) tea.Cmd {
	return tea.Sequence(e.init...)
}

func (e Editor) Update(ctx tea.Context, msg tea.Msg) (Editor, tea.Cmd) {
	var cmds []tea.Cmd
	var overwriteCursorBlink bool

	switch msg := msg.(type) {
	case ls.UpdateFileDiagnosticMsg:
		f := e.FileByName(msg.Name)
		if f == nil {
			return e, tea.Batch(cmds...)
		}
		f.SetDiagnostic(msg.Type, msg.Version, msg.Diagnostics)
		return e, tea.Batch(cmds...)
	case ls.UpdateAutocompletionMsg:
		f := e.FileByName(msg.Name)
		if f == nil {
			return e, tea.Batch(cmds...)
		}
		f.Autocomplete().SetCompletions(msg.Completions)
		return e, tea.Batch(cmds...)
	case ls.UpdateInlayHintMsg:
		f := e.FileByName(msg.Name)
		if f == nil {
			return e, tea.Batch(cmds...)
		}
		f.SetInlayHint(msg.Hints)
		return e, tea.Batch(cmds...)
	case ls.RefreshInlayHintMsg:
		// refresh inlay hints for all open files
		for _, f := range e.files {
			cmds = append(cmds, ls.GetInlayHint(f.Name(), f.Range()))
		}
		return e, tea.Batch(cmds...)
	case ls.UpdateDefinitionMsg:
		f := e.FileByName(msg.Name)
		if f == nil {
			return e, tea.Batch(cmds...)
		}
		cmds = append(cmds, f.SetDefinitions(msg.Definitions))
		return e, tea.Batch(cmds...)
	case file.OpenDirMsg:
		e.fileTree.Show()
		e.fileTree.Focus()

		if f := e.File(); f != nil {
			f.Blur()
		}

		if err := e.fileTree.Open(msg.Name); err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while opening directory %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("directory %s opened", msg.Name)))

		var wCmds []tea.Cmd
		if e.workspace != "" {
			wCmds = append(wCmds, ls.WorkspaceClosed(e.workspace))
		}
		e.workspace = msg.Name
		wCmds = append(wCmds, ls.WorkspaceOpened(msg.Name))
		return e, tea.Batch(append(cmds, tea.Sequence(wCmds...))...)
	case file.OpenFileMsg:
		cmd, err := e.OpenFile(msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while opening file %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s opened", msg.Name)))
		e.fileTree.Blur()
		e.SetFileByName(msg.Name)
		if msg.Position != nil {
			e.File().SetCursor(msg.Position.Row, msg.Position.Col)
		}
		cmds = append(cmds, e.File().Focus())
		return e, tea.Batch(cmds...)
	case file.SaveFileMsg:
		cmd, err := e.SaveFile(msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while saving file %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s saved", msg.Name)))
		return e, tea.Batch(cmds...)
	case file.CloseFileMsg:
		cmd, err := e.CloseFile(msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while closing file %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s closed", msg.Name)))
		return e, tea.Batch(cmds...)
	case file.FocusFileMsg:
		name := msg.Name
		if name == "" {
			name = e.File().Name()
		}
		e.SetFileByName(name)
		cmds = append(cmds, e.File().Focus())
		return e, tea.Batch(cmds...)
	case file.NewFileMsg:
		cmd, err := e.CreateFile(msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while creating file %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s created", msg.Name)))
		e.fileTree.Blur()
		e.SetFileByName(msg.Name)
		cmds = append(cmds, e.File().Focus())
		return e, tea.Batch(cmds...)
	case file.CloseAllMsg:
		var files []string
		for _, f := range e.files {
			if f.Dirty() {
				files = append(files, f.Name())
			}
		}
		if len(files) > 0 {
			return e, overlay.Open(NewCloseOverlay(files))
		}
		fileCmds := make([]tea.Cmd, len(e.files))
		for _, f := range e.files {
			fileCmds = append(fileCmds, file.CloseFile(f.Name()))
		}
		cmds = append(cmds, tea.Sequence(fileCmds...))
		return e, tea.Batch(cmds...)
	case searchbar.SearchMsg:
		f := e.File()
		if f != nil {
			results := f.Buffer().Search(msg.Term)
			cmds = append(cmds, searchbar.SearchResult(results))
			return e, tea.Batch(cmds...)
		}
		return e, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.ToggleFileTree) && e.focus:
			f := e.File()
			if e.fileTree.Visible() {
				e.fileTree.Blur()
				e.fileTree.Hide()
				if f != nil {
					cmds = append(cmds, f.Focus())
				}
			} else {
				e.fileTree.Focus()
				e.fileTree.Show()
				if f != nil {
					f.Blur()
				}
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.FocusFileTree) && e.focus:
			if e.fileTree.Visible() {
				f := e.File()
				if e.fileTree.Focused() {
					if f != nil {
						e.fileTree.Blur()
						cmds = append(cmds, f.Focus())
					}
				} else {
					e.fileTree.Focus()
					if f != nil {
						f.Blur()
					}
				}
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.NewFile) && e.focus:
			return e, overlay.Open(NewNewOverlay())
		}
	}

	var cmd tea.Cmd
	e.fileTree, cmd = e.fileTree.Update(ctx, msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if _, ok := msg.(tea.KeyMsg); ok && e.fileTree.Focused() {
		return e, tea.Batch(cmds...)
	}

	e.searchBar, cmd = e.searchBar.Update(ctx, msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if _, ok := msg.(tea.KeyMsg); ok && e.searchBar.Focused() {
		return e, tea.Batch(cmds...)
	}

	f := e.File()
	if f == nil {
		return e, tea.Batch(cmds...)
	}
	oldRow, oldCol := f.Cursor()

	switch msg := msg.(type) {
	case tea.PasteMsg:
		s := f.Selection()
		if s != nil {
			log.Println("paste msg", msg)
			f.Replace(s.Start.Row, s.Start.Col, s.End.Row, s.End.Col, []byte(msg))
			f.ResetMark()
		} else {
			f.Insert([]byte(msg))
		}
		return e, tea.Batch(cmds...)

	case file.PasteMsg:
		s := f.Selection()
		if s != nil {
			log.Println("paste msg", msg)
			f.Replace(s.Start.Row, s.Start.Col, s.End.Row, s.End.Col, msg)
			f.ResetMark()
		} else {
			f.Insert(msg)
		}
		return e, tea.Batch(cmds...)
	case file.CutMsg:
		s := buffer.Range(msg)
		f.DeleteRange(s.Start, s.End)
		f.ResetMark()
	case file.DeleteMsg:
		cmd, err := e.DeleteFile(f.Name())
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while deleting file %s: %s", f.Name(), err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s deleted", f.Name())))
	case file.RenameMsg:
		cmds = append(cmds, overlay.Open(NewRenameOverlay(f.Name())))
	case file.RenameFileMsg:
		cmd, err := e.RenameFile(f.Name(), msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while renamed file %s: %s", f.Name(), err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s renamed to %s", f.Name(), msg.Name)))
	case file.SetLanguageMsg:
		f.SetLanguage(msg.Language)

		f.ClearDiagnosticsByType(ls.DiagnosticTypeTreeSitter)

		if err := f.InitTree(); err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree: %s", err.Error())))
		}
		return e, tea.Batch(cmds...)
	case file.SaveMsg:
		if f.Dirty() {
			cmds = append(cmds, file.SaveFile(f.Name()))
		}
	case file.CloseMsg:
		if f.Dirty() {
			return e, overlay.Open(NewCloseOverlay([]string{f.Name()}))
		}
		cmds = append(cmds, file.CloseFile(f.Name()))
	case file.SelectMsg:
		f.SetMark(msg.FromRow, msg.FromCol)
		f.SetCursor(msg.ToRow, msg.ToCol)
	case file.ScrollMsg:
		f.SetCursor(msg.Row, msg.Col)
	case tea.MouseDownMsg:
		for _, z := range zone.GetPrefix(file.ZoneFileLinePrefix) {
			switch {
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseLeft):
				row, col := f.GetFileZoneCursorPos(tea.MouseEvent(msg), z)
				f.SetMark(row, col)
				f.SetCursor(row, col)
				return e, tea.Batch(cmds...)
			}
		}
	case tea.MouseUpMsg:
		for _, z := range zone.GetPrefix(file.ZoneFileLinePrefix) {
			switch {
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseLeft):
				row, col := f.GetFileZoneCursorPos(tea.MouseEvent(msg), z)
				f.SetCursor(row, col)
				if s := f.Selection(); s == nil || s.Zero() {
					f.ResetMark()
				}
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseLeft):
				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFilePrefix))
				e.SetFile(i)
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseRight):
				// TODO: open context menu?
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseMiddle):
				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFilePrefix))
				cmds = append(cmds, file.CloseFile(e.files[i].Name()))
				return e, tea.Batch(cmds...)

			case mouse.Matches(tea.MouseEvent(msg), ZoneFileLanguage, tea.MouseLeft):
				cmds = append(cmds, overlay.Open(NewSetLanguageOverlay()))
				return e, tea.Batch(cmds...)
			case mouse.Matches(tea.MouseEvent(msg), ZoneFileLineEnding, tea.MouseLeft):
				log.Println("file line ending zone")
				// cmds = append(cmds, overlay.Open(NewSetLineEndingOverlay()))
				return e, tea.Batch(cmds...)
			case mouse.Matches(tea.MouseEvent(msg), ZoneFileEncoding, tea.MouseLeft):
				log.Println("file encoding zone")
				// cmds = append(cmds, overlay.Open(NewSetEncodingOverlay()))
				return e, tea.Batch(cmds...)
			}
		}
	case tea.MouseMotionMsg:
		for _, z := range zone.GetPrefix(file.ZoneFileLinePrefix) {
			switch {
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseLeft):
				row, col := f.GetFileZoneCursorPos(tea.MouseEvent(msg), z)
				f.SetCursor(row, col)
				return e, tea.Batch(cmds...)
			}
		}
	case tea.MouseWheelMsg:
		for _, z := range zone.GetPrefix(file.ZoneFileLinePrefix) {
			switch {
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseWheelLeft), mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseWheelDown, tea.Shift):
				f.MoveCursorLeft(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseWheelRight), mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseWheelUp, tea.Shift):
				f.MoveCursorRight(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseWheelUp):
				f.MoveCursorUp(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(tea.MouseEvent(msg), z, tea.MouseWheelDown):
				f.MoveCursorDown(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			}
		}
	case tea.KeyMsg:
		if e.focus {
			switch {
			case key.Matches(msg, config.Keys.Editor.Autocomplete):
				row, col := f.Cursor()
				cmds = append(cmds, ls.GetAutocompletion(f.Name(), row, col))
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Cancel) && f.Autocomplete().Visible():
				f.Autocomplete().ClearCompletions()
			case key.Matches(msg, config.Keys.Editor.NextCompletion) && f.Autocomplete().Visible():
				f.Autocomplete().Next()
			case key.Matches(msg, config.Keys.Editor.PrevCompletion) && f.Autocomplete().Visible():
				f.Autocomplete().Previous()
			case key.Matches(msg, config.Keys.Editor.ApplyCompletion) && f.Autocomplete().Visible():
				completion := f.Autocomplete().Selected()
				if completion != nil {
					if completion.Text != "" {
						cmds = append(cmds, f.Insert([]byte(completion.Text)))
					} else if completion.Edit != nil {
						cmds = append(cmds, f.Replace(
							completion.Edit.Range.Start.Row,
							completion.Edit.Range.Start.Col,
							completion.Edit.Range.End.Row,
							completion.Edit.Range.End.Col,
							[]byte(completion.Edit.NewText)),
						)
					} else {
						cmds = append(cmds, f.Insert([]byte(completion.Label)))
					}
				}
				f.Autocomplete().ClearCompletions()
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.RefreshSyntaxHighlight):
				if err := f.InitTree(); err != nil {
					cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree sitter tree: %s", err.Error())))
				}
			case key.Matches(msg, config.Keys.Editor.ToggleTreeSitterDebug):
				e.ToggleTreeSitterDebug()
			case key.Matches(msg, config.Keys.Editor.DebugTreeSitterNodes):
				if f.Tree() == nil {
					cmds = append(cmds, notifications.Add("no tree available for this file"))
					return e, tea.Batch(cmds...)
				}
				buff, err := buffer.New(f.FileName()+".tree", bytes.NewReader([]byte(f.Tree().Print())), "utf-8", buffer.LineEndingLF, false)
				if err != nil {
					cmds = append(cmds, notifications.Add(fmt.Sprintf("error while opening tree.scm: %s", err.Error())))
					return e, tea.Batch(cmds...)
				}

				debugFile := file.NewFileWithBuffer(buff, file.ModeReadOnly)

				e.files = append(e.files, debugFile)
				e.activeFile = len(e.files) - 1
			case key.Matches(msg, config.Keys.Editor.ShowCurrentDiagnostic):
				f.ShowCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Editor.ShowDefinitions):
				cmds = append(cmds, f.ShowDefinitions())
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.OpenOutline):
				cmds = append(cmds, overlay.Open(NewOutlineOverlay(f)))
			case key.Matches(msg, config.Keys.Editor.Search):
				if !e.searchBar.Focused() {
					e.searchBar.Show()
					f.Blur()
					cmds = append(cmds, e.searchBar.Focus())
				}
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.NextFile):
				if e.activeFile < len(e.files)-1 {
					e.activeFile++
					f.Blur()
					cmds = append(cmds, e.files[e.activeFile].Focus())
				}
			case key.Matches(msg, config.Keys.Editor.PrevFile):
				if e.activeFile > 0 {
					e.activeFile--
					f.Blur()
					cmds = append(cmds, e.files[e.activeFile].Focus())
				}
			case key.Matches(msg, config.Keys.Editor.CloseFile):
				cmds = append(cmds, file.Close)

			case key.Matches(msg, config.Keys.Editor.DeleteFile):
				cmds = append(cmds, overlay.Open(NewDeleteOverlay()))
			case key.Matches(msg, config.Keys.Editor.RenameFile):
				cmds = append(cmds, overlay.Open(NewRenameOverlay(f.Name())))
			case key.Matches(msg, config.Keys.Editor.LineUp):
				f.MoveCursorUp(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.LineDown):
				f.MoveCursorDown(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.CharacterLeft):
				f.MoveCursorLeft(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.CharacterRight):
				f.MoveCursorRight(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.WordUp):
				f.MoveCursorWordUp()
			case key.Matches(msg, config.Keys.Editor.WordDown):
				f.MoveCursorWordDown()
			case key.Matches(msg, config.Keys.Editor.WordLeft):
				f.SetCursor(f.NextWordLeft())
			case key.Matches(msg, config.Keys.Editor.WordRight):
				f.SetCursor(f.NextWordRight())
			case key.Matches(msg, config.Keys.Editor.PageUp):
				f.MoveCursorUp(pageSize)
			case key.Matches(msg, config.Keys.Editor.PageDown):
				f.MoveCursorDown(pageSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.LineStart):
				f.SetCursor(-1, 0)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.LineEnd):
				cursorRow, _ := f.Cursor()
				f.SetCursor(-1, f.Buffer().LineLen(cursorRow))
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.FileStart):
				f.SetCursor(0, -1)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.FileEnd):
				f.SetCursor(f.Buffer().LinesLen(), -1)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Copy):
				selBytes := f.SelectionBytes()
				if len(selBytes) > 0 {
					cmds = append(cmds, file.Copy(selBytes))
				}
			case key.Matches(msg, config.Keys.Editor.Paste):
				cmds = append(cmds, file.Paste)
			case key.Matches(msg, config.Keys.Editor.Cut):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, file.Cut(*s, f.SelectionBytes()))
				}
			case key.Matches(msg, config.Keys.Editor.SelectLeft):
				f.SelectLeft(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectRight):
				f.SelectRight(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectUp):
				f.SelectUp(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectDown):
				f.SelectDown(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectAll):
				f.SelectAll()

			case key.Matches(msg, config.Keys.Editor.SaveFile):
				cmds = append(cmds, file.SaveFile(f.Name()))
			case key.Matches(msg, config.Keys.Editor.Tab):
				cmds = append(cmds, f.InsertRunes([]rune{'\t'}))
			case key.Matches(msg, config.Keys.Editor.RemoveTab):
				cmds = append(cmds, f.RemoveTab())
			case key.Matches(msg, config.Keys.Editor.Newline):
				f.ResetMark()
				cmds = append(cmds, f.InsertNewLine(), f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.DeleteRight):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteAfter(1))
				}
			case key.Matches(msg, config.Keys.Editor.DeleteLeft):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					row, col := f.Cursor()
					toDelete := f.Buffer().Line(row).RuneBytes(col - 1)
					cmds = append(cmds, f.DeleteBefore(1))
					if lang := f.Language(); lang != nil && len(lang.Config.AutoPairs) > 0 {
						row, col = f.Cursor()

						fileTree := f.Tree()
						if fileTree != nil {
							tree := fileTree.FindTree(buffer.Position{
								Row: row,
								Col: col,
							})
							if tree != nil {
								node := tree.Tree.RootNode().DescendantForRange(sitter.Point{
									Row:    uint32(row),
									Column: uint32(col),
								},
									sitter.Point{
										Row:    uint32(row),
										Column: uint32(col),
									},
								)
								if node != nil && node.Type() == "string" {
									log.Println("IN STRING")
								}
							}
						}

						for _, pair := range lang.Config.AutoPairs {
							if string(toDelete) != pair.Open {
								continue
							}
							closeWidth := runewidth.StringWidth(pair.Close)
							behindCursor := f.Buffer().BytesRange(
								buffer.Position{
									Row: row,
									Col: col,
								},
								buffer.Position{
									Row: row,
									Col: col + closeWidth,
								},
							)
							if string(behindCursor) == pair.Close {
								cmds = append(cmds, f.Replace(row, col, row, col+closeWidth, nil))
								break
							}
						}
					}
				}
			case key.Matches(msg, config.Keys.Editor.DuplicateLine):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.Insert(f.SelectionBytes()))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DuplicateLine())
				}
			case key.Matches(msg, config.Keys.Editor.DeleteWordLeft):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteWordLeft())
				}
			case key.Matches(msg, config.Keys.Editor.DeleteWordRight):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteWordRight())
				}
			case key.Matches(msg, config.Keys.Editor.DeleteLine):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteLine())
				}
			case key.Matches(msg, config.Keys.Editor.ToggleComment):
				cmds = append(cmds, f.ToggleComment())
				overwriteCursorBlink = true
			case key.Matches(msg, config.Keys.Debug):
				log.Println("DEBUG")

			default:
				if msg.Mod == 0 {
					text := []byte(string(msg.Rune))
					if s := f.Selection(); s != nil {
						cmds = append(cmds, f.Replace(s.Start.Row, s.Start.Col, s.End.Row, s.End.Col, text))
						f.ResetMark()
					} else {
						cmds = append(cmds, f.Insert(text))
					}

					cmds = append(cmds, f.Autocomplete().Update())

					// handle auto pairs
					if lang := f.Language(); lang != nil && len(lang.Config.AutoPairs) > 0 {
						for _, pair := range lang.Config.AutoPairs {
							if string(msg.Rune) == pair.Open {
								row, col := f.Cursor()
								cmds = append(cmds, f.InsertAt(row, col+runewidth.StringWidth(pair.Open), []byte(pair.Close)))
								break
							}
						}
					}
				}
			}
		}
	}

	if cmd = f.UpdateCursor(ctx, msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	newRow, newCol := f.Cursor()
	if oldRow != newRow || oldCol != newCol || overwriteCursorBlink {
		f.SetCursorBlink(false)
		cmds = append(cmds, f.CursorBlinkCmd())
	}

	return e, tea.Batch(cmds...)
}

func (e *Editor) View(ctx tea.Context, height int) string {
	width, _ := ctx.WindowSize()

	var fileTree string
	if e.fileTree.Visible() {
		fileTree = e.fileTree.View(height)
		width -= lipgloss.Width(fileTree)
	}

	f := e.File()
	if f == nil {
		width -= config.Theme.UI.FileView.EmptyStyle.GetHorizontalBorderSize()
		height -= config.Theme.UI.FileView.EmptyStyle.GetVerticalBorderSize()

		code := config.Theme.UI.FileView.EmptyStyle.
			Width(width).
			Height(height).
			Render(fmt.Sprintf("No file open.\n\nPress '%s' to open a file.", config.Keys.Editor.OpenFile.Help().Key))

		if fileTree == "" {
			return code
		}
		code = config.Theme.UI.FileView.BorderStyle.Render(code)
		return lipgloss.JoinHorizontal(lipgloss.Top, fileTree, code)
	}

	var searchBar string
	if e.searchBar.Visible() {
		searchBar = e.searchBar.View(ctx)
		height -= lipgloss.Height(searchBar)
	}

	editor := f.View(ctx, width, height, e.fileTree.Visible(), e.treeSitterDebug)

	if searchBar != "" {
		editor = lipgloss.JoinVertical(lipgloss.Left, searchBar, editor)
	}

	if fileTree == "" {
		return editor
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, fileTree, editor)
}

func (e *Editor) refreshActiveFileOffset(width int, fileNames []string) {
	if e.activeFileOffset == 0 {
		return
	}

	filesWidth := lipgloss.Width(strings.Join(fileNames[:e.activeFile], ""))
	if filesWidth < width {
		e.activeFileOffset = 0
	}

	if filesWidth > width {
		e.activeFileOffset = e.activeFile
	}

	for i := e.activeFileOffset; i > 0; i-- {
		filesWidth = lipgloss.Width(strings.Join(fileNames[i:e.activeFile], ""))
		if filesWidth < width {
			e.activeFileOffset = i
			break
		}
	}
}

func (e *Editor) FileTabsView(width int) string {
	var fileNames []string
	for i, f := range e.files {
		var languageName string
		if f.Language() != nil {
			languageName = f.Language().Name
		}
		icon := config.Theme.Icons.FileIcon(languageName).Render()

		fileName := clampString(f.FileName(), 16)
		fileName = fmt.Sprintf("%s %s", icon, fileName)
		if f.Dirty() {
			fileName += "*"
		} else {
			fileName += " "
		}

		style := config.Theme.UI.AppBar.Files.FileStyle
		if i == e.activeFile {
			style = config.Theme.UI.AppBar.Files.SelectedFileStyle
		}
		fileNames = append(fileNames, zone.Mark(fmt.Sprintf("file:%d", i), style.Render(fileName)))
	}

	if config.Gopad.FileView.OpenFilesWrap {
		var fileTabs string
		var line string
		for _, fileName := range fileNames {
			fileNameWidth := lipgloss.Width(fileName)
			if fileNameWidth+lipgloss.Width(line) > width {
				fileTabs += line + "\n"
				line = ""
			}
			line += fileName
		}
		if line != "" {
			fileTabs += line + "\n"
		}
		return strings.TrimRight(fileTabs, "\n")
	}

	e.refreshActiveFileOffset(width, fileNames)

	return config.Theme.UI.AppBar.Files.Style.Render(fileNames[e.activeFileOffset:]...)
}

func clampString(s string, length int) string {
	if len(s) > length {
		return s[:length-1] + "…"
	}
	return s
}
