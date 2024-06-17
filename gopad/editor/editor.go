package editor

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/filetree"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/bubbles/searchbar"
)

const (
	moveSize = 1
	pageSize = 10
)

var fileIconByFileNameFunc = func(name string) rune {
	language := GetLanguageByFilename(name)
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
				return Scroll(result.RowStart, result.ColStart)
			},
			FocusFile(""),
		),
		fileTree:  config.NewFileTree(OpenFile, fileIconByFileNameFunc),
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

	if file := editor.File(); file != nil {
		cmds = append(cmds, file.Focus())
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
	files            []*File
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
	if slices.ContainsFunc(e.files, func(b *File) bool {
		return b.Name() == name
	}) {
		return nil, nil
	}

	buff, err := buffer.New(name, bytes.NewReader(nil), "utf-8", buffer.LineEndingLF, false)
	if err != nil {
		return nil, err
	}
	file := NewFileWithBuffer(buff, FileModeWrite)

	e.files = append(e.files, file)

	cmds := []tea.Cmd{
		tea.Sequence(
			ls.FileCreated(file.Name(), file.buffer.Bytes()),
			ls.FileOpened(file.Name(), file.buffer.Version(), file.buffer.Bytes()),
		),
	}

	if err = file.InitTree(); err != nil {
		cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree sitter tree: %s", err.Error())))
	}
	cmds = append(cmds, ls.GetInlayHint(file.Name(), file.Range()))

	return tea.Batch(cmds...), nil
}

func (e *Editor) OpenFile(name string) (tea.Cmd, error) {
	if slices.ContainsFunc(e.files, func(b *File) bool {
		return b.Name() == name
	}) {
		return nil, nil
	}

	file, err := NewFileFromName(name)
	if err != nil {
		return nil, err
	}
	e.files = append(e.files, file)

	cmds := []tea.Cmd{
		ls.FileOpened(file.Name(), file.buffer.Version(), file.buffer.Bytes()),
	}

	if err = file.InitTree(); err != nil {
		cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree sitter tree: %s", err.Error())))
	}
	cmds = append(cmds, ls.GetInlayHint(file.Name(), file.Range()))

	return tea.Batch(cmds...), nil
}

func (e *Editor) SaveFile(name string) (tea.Cmd, error) {
	file := e.FileByName(name)
	if file == nil {
		return nil, nil
	}
	if err := file.buffer.Save(); err != nil {
		return nil, err
	}

	return ls.FileSaved(file.Name(), file.buffer.Bytes()), nil
}

func (e *Editor) RenameFile(oldName string, newName string) (tea.Cmd, error) {
	if !filepath.IsAbs(newName) {
		newName = filepath.Join(e.workspace, newName)
		newName, _ = filepath.Abs(newName)
	}
	file := e.FileByName(oldName)
	if file == nil {
		return nil, nil
	}
	if err := file.buffer.Rename(newName); err != nil {
		return nil, err
	}

	return ls.FileRenamed(file.Name(), newName), nil
}

func (e *Editor) CloseFile(name string) (tea.Cmd, error) {
	index := slices.IndexFunc(e.files, func(file *File) bool {
		return file.Name() == name
	})
	if index == -1 {
		return nil, nil
	}

	file := e.files[index]
	e.files = slices.Delete(e.files, index, index+1)
	e.activeFile = min(e.activeFile, len(e.files)-1)
	if len(e.files) > 0 {
		e.files[e.activeFile].Focus()
	} else {
		e.fileTree.Focus()
	}

	return ls.FileClosed(file.Name()), nil
}

func (e *Editor) DeleteFile(name string) (tea.Cmd, error) {
	index := slices.IndexFunc(e.files, func(file *File) bool {
		return file.Name() == name
	})
	if index == -1 {
		return nil, nil
	}

	file := e.files[index]
	if err := file.buffer.Delete(); err != nil {
		return nil, err
	}

	e.files = slices.Delete(e.files, e.activeFile, e.activeFile+1)
	e.activeFile = min(e.activeFile, len(e.files)-1)
	if len(e.files) > 0 {
		e.files[e.activeFile].Focus()
	} else {
		e.fileTree.Focus()
	}

	return ls.FileDeleted(file.Name()), nil
}

func (e *Editor) File() *File {
	if len(e.files) == 0 {
		return nil
	}
	return e.files[e.activeFile]
}

func (e *Editor) SetFile(index int) {
	e.activeFile = index
}

func (e *Editor) SetFileByName(name string) {
	for i, file := range e.files {
		if file.Name() == name {
			e.activeFile = i
			return
		}
	}
}

func (e *Editor) FileByName(name string) *File {
	for _, file := range e.files {
		if file.Name() == name {
			return file
		}
	}
	return nil
}

func (e *Editor) HasChanges() bool {
	for _, file := range e.files {
		if file.Dirty() {
			return true
		}
	}
	return false
}

func (e *Editor) ToggleTreeSitterDebug() {
	e.treeSitterDebug = !e.treeSitterDebug
}

func (e Editor) Init() tea.Cmd {
	return tea.Sequence(e.init...)
}

func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd) {
	var cmds []tea.Cmd
	var overwriteCursorBlink bool

	switch msg := msg.(type) {
	case ls.UpdateFileDiagnosticMsg:
		file := e.FileByName(msg.Name)
		if file == nil {
			return e, tea.Batch(cmds...)
		}
		file.SetDiagnostic(msg.Type, msg.Version, msg.Diagnostics)
		return e, tea.Batch(cmds...)
	case ls.UpdateAutocompletionMsg:
		log.Println("update autocompletions", msg.Name, msg.Completions)
		file := e.FileByName(msg.Name)
		if file == nil {
			return e, tea.Batch(cmds...)
		}
		file.autocomplete.SetCompletions(msg.Completions)
		return e, tea.Batch(cmds...)
	case ls.UpdateInlayHintMsg:
		file := e.FileByName(msg.Name)
		if file == nil {
			return e, tea.Batch(cmds...)
		}
		file.SetInlayHint(msg.Hints)
		return e, tea.Batch(cmds...)
	case ls.RefreshInlayHintMsg:
		// refresh inlay hints for all open files
		for _, file := range e.files {
			cmds = append(cmds, ls.GetInlayHint(file.Name(), file.Range()))
		}
		return e, tea.Batch(cmds...)
	case openDirMsg:
		e.fileTree.Show()
		e.fileTree.Focus()

		if file := e.File(); file != nil {
			file.Blur()
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
	case openFileMsg:
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
		cmds = append(cmds, e.File().Focus())
		return e, tea.Batch(cmds...)
	case saveFileMsg:
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
	case closeFileMsg:
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
	case focusFileMsg:
		name := msg.Name
		if name == "" {
			name = e.File().Name()
		}
		e.SetFileByName(name)
		cmds = append(cmds, e.File().Focus())
		return e, tea.Batch(cmds...)
	case newFileMsg:
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
	case closeAllMsg:
		var files []string
		for _, file := range e.files {
			if file.Dirty() {
				files = append(files, file.Name())
			}
		}
		if len(files) > 0 {
			return e, overlay.Open(NewCloseOverlay(files))
		}
		fileCmds := make([]tea.Cmd, len(e.files))
		for _, file := range e.files {
			fileCmds = append(fileCmds, CloseFile(file.Name()))
		}
		cmds = append(cmds, tea.Sequence(fileCmds...))
		return e, tea.Batch(cmds...)
	case searchbar.SearchMsg:
		file := e.File()
		if file != nil {
			results := file.buffer.Search(msg.Term)
			cmds = append(cmds, searchbar.SearchResult(results))
			return e, tea.Batch(cmds...)
		}
		return e, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.ToggleFileTree) && e.focus:
			file := e.File()
			if e.fileTree.Visible() {
				e.fileTree.Blur()
				e.fileTree.Hide()
				if file != nil {
					cmds = append(cmds, file.Focus())
				}
			} else {
				e.fileTree.Focus()
				e.fileTree.Show()
				if file != nil {
					file.Blur()
				}
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.FocusFileTree) && e.focus:
			if e.fileTree.Visible() {
				file := e.File()
				if e.fileTree.Focused() {
					e.fileTree.Blur()
					if file != nil {
						cmds = append(cmds, file.Focus())
					}
				} else {
					e.fileTree.Focus()
					if file != nil {
						file.Blur()
					}
				}
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.NewFile) && e.focus:
			return e, overlay.Open(NewNewOverlay())
		}
	}

	var cmd tea.Cmd
	e.fileTree, cmd = e.fileTree.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if _, ok := msg.(tea.KeyMsg); ok && e.fileTree.Focused() {
		return e, tea.Batch(cmds...)
	}

	e.searchBar, cmd = e.searchBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if _, ok := msg.(tea.KeyMsg); ok && e.searchBar.Focused() {
		return e, tea.Batch(cmds...)
	}

	file := e.File()
	if file == nil {
		return e, nil
	}
	oldRow, oldCol := file.Cursor()

	switch msg := msg.(type) {
	case pasteMsg:
		s := file.Selection()
		if s != nil {
			log.Println("paste msg", msg)
			file.Replace(s.Start.Row, s.Start.Col, s.End.Row, s.End.Col, msg)
			file.ResetMark()
		} else {
			file.Insert(msg)
		}
	case cutMsg:
		s := buffer.Range(msg)
		file.DeleteRange(s.Start, s.End)
		file.ResetMark()
	case deleteMsg:
		cmd, err := e.DeleteFile(file.Name())
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while deleting file %s: %s", file.Name(), err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s deleted", file.Name())))
	case renameMsg:
		cmds = append(cmds, overlay.Open(NewRenameOverlay(file.Name())))
	case renameFileMsg:
		cmd, err := e.RenameFile(file.Name(), msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while renamed file %s: %s", file.Name(), err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s renamed to %s", file.Name(), msg.Name)))
	case setLanguageMsg:
		file.SetLanguage(msg.Language)

		file.ClearDiagnosticsByType(ls.DiagnosticTypeTreeSitter)

		if err := file.InitTree(); err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree: %s", err.Error())))
		}
		return e, tea.Batch(cmds...)
	case saveMsg:
		if file.Dirty() {
			cmds = append(cmds, SaveFile(file.Name()))
		}
	case closeMsg:
		if file.Dirty() {
			return e, overlay.Open(NewCloseOverlay([]string{file.Name()}))
		}
		cmds = append(cmds, CloseFile(file.Name()))
	case selectMsg:
		file.SetMark(msg.FromRow, msg.FromCol)
		file.SetCursor(msg.ToRow, msg.ToCol)
	case scrollMsg:
		file.SetCursor(msg.Row, msg.Col)
	case tea.KeyMsg:
		if e.focus {
			switch {
			case key.Matches(msg, config.Keys.Editor.Autocomplete):
				row, col := file.Cursor()
				cmds = append(cmds, ls.GetAutocompletion(file.Name(), row, col))
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Cancel) && file.autocomplete.Visible():
				file.autocomplete.ClearCompletions()
			case key.Matches(msg, config.Keys.Editor.NextCompletion) && file.autocomplete.Visible():
				file.autocomplete.Next()
			case key.Matches(msg, config.Keys.Editor.PrevCompletion) && file.autocomplete.Visible():
				file.autocomplete.Previous()
			case key.Matches(msg, config.Keys.Editor.ApplyCompletion) && file.autocomplete.Visible():
				completion := file.autocomplete.Selected()
				if completion.Text != "" {
					cmds = append(cmds, file.Insert([]byte(completion.Text)))
				} else if completion.Edit != nil {
					cmds = append(cmds, file.Replace(
						completion.Edit.Range.Start.Row,
						completion.Edit.Range.Start.Col,
						completion.Edit.Range.End.Row,
						completion.Edit.Range.End.Col,
						[]byte(completion.Edit.NewText)),
					)
				} else {
					cmds = append(cmds, file.Insert([]byte(completion.Label)))
				}
				file.autocomplete.ClearCompletions()

			case key.Matches(msg, config.Keys.Editor.RefreshSyntaxHighlight):
				if err := file.InitTree(); err != nil {
					cmds = append(cmds, notifications.Add(fmt.Sprintf("error refreshing tree sitter tree: %s", err.Error())))
				}
			case key.Matches(msg, config.Keys.Editor.ToggleTreeSitterDebug):
				e.ToggleTreeSitterDebug()
			case key.Matches(msg, config.Keys.Editor.DebugTreeSitterNodes):
				buff, err := buffer.New("tree.scm", bytes.NewReader([]byte(file.Tree().Print())), "utf-8", buffer.LineEndingLF, false)
				if err != nil {
					cmds = append(cmds, notifications.Add(fmt.Sprintf("error while opening tree.scm: %s", err.Error())))
					return e, tea.Batch(cmds...)
				}

				debugFile := NewFileWithBuffer(buff, FileModeReadOnly)

				e.files = append(e.files, debugFile)
				e.activeFile = len(e.files) - 1
			case key.Matches(msg, config.Keys.Editor.ShowCurrentDiagnostic):
				file.ShowCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Editor.OpenOutline):
				cmds = append(cmds, overlay.Open(NewOutlineOverlay(file)))
			case key.Matches(msg, config.Keys.Editor.Search):
				if !e.searchBar.Focused() {
					e.searchBar.Show()
					file.Blur()
					cmds = append(cmds, e.searchBar.Focus())
				}
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.NextFile):
				if e.activeFile < len(e.files)-1 {
					e.activeFile++
					file.Blur()
					cmds = append(cmds, e.files[e.activeFile].Focus())
				}
			case key.Matches(msg, config.Keys.Editor.PrevFile):
				if e.activeFile > 0 {
					e.activeFile--
					file.Blur()
					cmds = append(cmds, e.files[e.activeFile].Focus())
				}
			case key.Matches(msg, config.Keys.Editor.CloseFile):
				cmds = append(cmds, Close)

			case key.Matches(msg, config.Keys.Editor.DeleteFile):
				cmds = append(cmds, overlay.Open(NewDeleteOverlay()))
			case key.Matches(msg, config.Keys.Editor.RenameFile):
				cmds = append(cmds, overlay.Open(NewRenameOverlay(file.Name())))
			case key.Matches(msg, config.Keys.Editor.LineUp):
				file.MoveCursorUp(moveSize)
			case key.Matches(msg, config.Keys.Editor.LineDown):
				file.MoveCursorDown(moveSize)
			case key.Matches(msg, config.Keys.Editor.CharacterLeft):
				file.MoveCursorLeft(moveSize)
			case key.Matches(msg, config.Keys.Editor.CharacterRight):
				file.MoveCursorRight(moveSize)
			case key.Matches(msg, config.Keys.Editor.WordUp):
				file.MoveCursorWordUp()
			case key.Matches(msg, config.Keys.Editor.WordDown):
				file.MoveCursorWordDown()
			case key.Matches(msg, config.Keys.Editor.WordLeft):
				file.SetCursor(file.NextWordLeft())
			case key.Matches(msg, config.Keys.Editor.WordRight):
				file.SetCursor(file.NextWordRight())
			case key.Matches(msg, config.Keys.Editor.PageUp):
				file.MoveCursorUp(pageSize)
			case key.Matches(msg, config.Keys.Editor.PageDown):
				file.MoveCursorDown(pageSize)
			case key.Matches(msg, config.Keys.Editor.LineStart):
				file.SetCursor(-1, 0)
			case key.Matches(msg, config.Keys.Editor.LineEnd):
				cursorRow, _ := file.Cursor()
				file.SetCursor(-1, file.buffer.LineLen(cursorRow))
			case key.Matches(msg, config.Keys.Editor.FileStart):
				file.SetCursor(0, -1)
			case key.Matches(msg, config.Keys.Editor.FileEnd):
				file.SetCursor(file.buffer.LinesLen(), -1)

			case key.Matches(msg, config.Keys.Editor.Copy):
				selBytes := file.SelectionBytes()
				if len(selBytes) > 0 {
					cmds = append(cmds, Copy(selBytes))
				}
			case key.Matches(msg, config.Keys.Editor.Paste):
				cmds = append(cmds, Paste)
			case key.Matches(msg, config.Keys.Editor.Cut):
				s := file.Selection()
				if s != nil {
					cmds = append(cmds, Cut(*s, file.SelectionBytes()))
				}
			case key.Matches(msg, config.Keys.Editor.SelectLeft):
				file.SelectLeft(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectRight):
				file.SelectRight(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectUp):
				file.SelectUp(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectDown):
				file.SelectDown(moveSize)
			case key.Matches(msg, config.Keys.Editor.SelectAll):
				file.SelectAll()

			case key.Matches(msg, config.Keys.Editor.SaveFile):
				cmds = append(cmds, SaveFile(file.Name()))
			case key.Matches(msg, config.Keys.Editor.Tab):
				cmds = append(cmds, file.InsertRunes([]rune{'\t'}))
			case key.Matches(msg, config.Keys.Editor.RemoveTab):
				cmds = append(cmds, file.RemoveTab())
			case key.Matches(msg, config.Keys.Editor.Newline):
				file.ResetMark()
				cmds = append(cmds, file.InsertNewLine())
			case key.Matches(msg, config.Keys.Editor.DeleteRight):
				s := file.Selection()
				if s != nil {
					cmds = append(cmds, file.DeleteRange(s.Start, s.End))
					file.ResetMark()
				} else {
					cmds = append(cmds, file.DeleteAfter(1))
				}
			case key.Matches(msg, config.Keys.Editor.DeleteLeft):
				s := file.Selection()
				if s != nil {
					cmds = append(cmds, file.DeleteRange(s.Start, s.End))
					file.ResetMark()
				} else {
					row, col := file.Cursor()
					toDelete := file.buffer.Line(row).RuneBytes(col - 1)
					cmds = append(cmds, file.DeleteBefore(1))
					if lang := file.Language(); lang != nil && len(lang.Config.AutoPairs) > 0 {
						row, col = file.Cursor()

						fileTree := file.Tree()
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
							behindCursor := file.buffer.BytesRange(
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
								cmds = append(cmds, file.Replace(row, col, row, col+closeWidth, nil))
								break
							}
						}
					}
				}
			case key.Matches(msg, config.Keys.Editor.DuplicateLine):
				s := file.Selection()
				if s != nil {
					cmds = append(cmds, file.Insert(file.SelectionBytes()))
					file.ResetMark()
				} else {
					cmds = append(cmds, file.DuplicateLine())
				}
			case key.Matches(msg, config.Keys.Editor.DeleteWordLeft):
				s := file.Selection()
				if s != nil {
					cmds = append(cmds, file.DeleteRange(s.Start, s.End))
					file.ResetMark()
				} else {
					cmds = append(cmds, file.DeleteWordLeft())
				}
			case key.Matches(msg, config.Keys.Editor.DeleteWordRight):
				s := file.Selection()
				if s != nil {
					cmds = append(cmds, file.DeleteRange(s.Start, s.End))
					file.ResetMark()
				} else {
					cmds = append(cmds, file.DeleteWordRight())
				}
			case key.Matches(msg, config.Keys.Editor.DeleteLine):
				s := file.Selection()
				if s != nil {
					cmds = append(cmds, file.DeleteRange(s.Start, s.End))
					file.ResetMark()
				} else {
					cmds = append(cmds, file.DeleteLine())
				}
			case key.Matches(msg, config.Keys.Editor.ToggleComment):
				cmds = append(cmds, file.ToggleComment())
				overwriteCursorBlink = true
			case key.Matches(msg, config.Keys.Editor.Debug):
				log.Println("DEBUG")
				cmds = append(cmds, ls.GetInlayHint(file.Name(), buffer.Range{
					Start: buffer.Position{
						Row: 0,
						Col: 0,
					},
					End: buffer.Position{
						Row: file.buffer.LinesLen(),
						Col: file.buffer.LineLen(max(file.buffer.LinesLen()-1, 0)),
					},
				}))
				return e, tea.Batch(cmds...)

			default:
				if msg.Alt {
					break
				}
				cmds = append(cmds, file.InsertRunes(msg.Runes))
				if file.autocomplete.Visible() {
					row, col := file.Cursor()
					cmds = append(cmds, ls.GetAutocompletion(file.Name(), row, col))
				}

				// handle auto pairs
				if lang := file.Language(); lang != nil && len(lang.Config.AutoPairs) > 0 {
					for _, pair := range lang.Config.AutoPairs {
						if string(msg.Runes) == pair.Open {
							row, col := file.Cursor()
							cmds = append(cmds, file.InsertAt(row, col+runewidth.StringWidth(pair.Open), []byte(pair.Close)))
							break
						}
					}
				}
			}
		}
	}

	if cmd = file.UpdateCursor(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	newRow, newCol := file.Cursor()
	if oldRow != newRow || oldCol != newCol || overwriteCursorBlink {
		file.SetCursorBlink(false)
		cmds = append(cmds, file.CursorBlinkCmd())
	}

	return e, tea.Batch(cmds...)
}

func (e *Editor) View(width int, height int) string {
	var fileTree string
	if e.fileTree.Visible() {
		fileTree = e.fileTree.View(height)
		width -= lipgloss.Width(fileTree)
	}

	file := e.File()
	if file == nil {
		width -= config.Theme.Editor.EmptyStyle.GetHorizontalBorderSize()
		height -= config.Theme.Editor.EmptyStyle.GetVerticalBorderSize()

		code := config.Theme.Editor.EmptyStyle.
			Width(width).
			Height(height).
			Render(fmt.Sprintf("No file open.\n\nPress '%s' to open a file.", config.Keys.Editor.OpenFile.Help().Key))

		if fileTree == "" {
			return code
		}
		code = config.Theme.Editor.CodeBorderStyle.Render(code)
		return lipgloss.JoinHorizontal(lipgloss.Top, fileTree, code)
	}

	var searchBar string
	if e.searchBar.Visible() {
		searchBar = e.searchBar.View()
		height -= lipgloss.Height(searchBar)
	}

	editor := file.View(width, height, e.fileTree.Visible(), e.treeSitterDebug)

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
	for path, file := range e.files {
		var languageName string
		if file.language != nil {
			languageName = file.language.Name
		}
		icon := config.Theme.Icons.FileIcon(languageName)

		fileName := clampString(file.FileName(), 16)
		fileName = fmt.Sprintf("%c %s", icon, fileName)
		if file.Dirty() {
			fileName += "*"
		} else {
			fileName += " "
		}

		style := config.Theme.Editor.FileStyle
		if path == e.activeFile {
			style = config.Theme.Editor.FileSelectedStyle
		}
		fileNames = append(fileNames, style.Render(fileName))
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

	return strings.Join(fileNames[e.activeFileOffset:], "")
}

func clampString(s string, length int) string {
	if len(s) > length {
		return s[:length-1] + "…"
	}
	return s
}
