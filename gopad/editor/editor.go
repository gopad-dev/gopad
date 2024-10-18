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

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/bubblezone"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/internal/bubbles/key"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/editormsg"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/gopad/editor/filetree"
	"go.gopad.dev/gopad/gopad/editor/searchbar"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
)

const (
	moveSize = 1
	pageSize = 10

	ZoneFileLanguage   = "file.language"
	ZoneFileLineEnding = "file.lineEnding"
	ZoneFileEncoding   = "file.encoding"
	ZoneFileGoTo       = "file.goto"
	ZoneFilePrefix     = "file:"
)

func NewEditor(workspace string, args []string) (Editor, error) {
	e := Editor{
		args:      args,
		searchBar: searchbar.New(),
		fileTree:  filetree.New(),
		workspace: workspace,
	}

	if workspace != "" {
		if err := e.fileTree.Open(workspace); err != nil {
			return Editor{}, fmt.Errorf("failed to init file tree: %w", err)
		}
	}

	return e, nil
}

type Editor struct {
	fileTree         filetree.Model
	args             []string
	workspace        string
	searchBar        searchbar.Model
	files            []*file.File
	activeFile       int
	activeFileOffset int
	focus            bool
	treeSitterDebug  bool
}

func (e Editor) Init() (Editor, tea.Cmd) {
	var cmds []tea.Cmd

	if e.workspace != "" {
		e.fileTree.Show()
		cmds = append(cmds, ls.WorkspaceOpened(e.workspace))
	}

	for _, arg := range e.args {
		stat, err := os.Stat(arg)
		if errors.Is(err, os.ErrNotExist) {
			cmd, err := e.CreateFile(arg)
			if err != nil {
				return e, notifications.Addf("error while creating file %s: %s", arg, err)
			}
			cmds = append(cmds, cmd)
			continue
		}
		if err != nil {
			return e, notifications.Addf("error while checking file %s: %s", arg, err)
		}

		if stat.IsDir() {
			continue
		}
		cmd, err := e.OpenFile(arg)
		if err != nil {
			return e, notifications.Addf("error while opening file %s: %s", arg, err)
		}
		cmds = append(cmds, cmd)
	}

	if f := e.File(); f != nil {
		cmds = append(cmds, f.Focus())
	} else {
		e.fileTree.Focus()
	}

	return e, tea.Batch(cmds...)
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

	f := e.File()
	if f != nil {
		f.Blur()
	}
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
			ls.GetInlayHint(f.Name(), f.Version(), f.Range()),
		),
	}

	if cmd := f.InitTree(); cmd != nil {
		cmds = append(cmds, cmd)
	}

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
		ls.GetInlayHint(f.Name(), f.Version(), f.Range()),
	}

	if cmd := f.InitTree(); cmd != nil {
		cmds = append(cmds, cmd)
	}

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

func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd) {
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
		f.SetInlayHint(msg.Version, msg.Hints)
		return e, tea.Batch(cmds...)
	case file.UpdateMatchesMsg:
		f := e.FileByName(msg.Name)
		if f == nil {
			return e, tea.Batch(cmds...)
		}
		f.SetMatches(msg.Version, msg.Matches)
		return e, tea.Batch(cmds...)
	case ls.RefreshInlayHintMsg:
		// refresh inlay hints for all open files
		for _, f := range e.files {
			cmds = append(cmds, ls.GetInlayHint(f.Name(), f.Version(), f.Range()))
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
		cmds = append(cmds,
			notifications.Add(fmt.Sprintf("file %s opened", msg.Name)),
			editormsg.Focus(editormsg.ModelFile),
		)
		e.SetFileByName(msg.Name)
		if msg.Position != nil {
			e.File().SetCursor(msg.Position.Row, msg.Position.Col)
		}
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
	case file.NewFileMsg:
		cmd, err := e.CreateFile(msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while creating file %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds,
			notifications.Add(fmt.Sprintf("file %s created", msg.Name)),
			editormsg.Focus(editormsg.ModelFile),
		)
		e.SetFileByName(msg.Name)
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
			cmds = append(cmds, searchbar.ExecSearch(msg.Term, f.Buffer().Copy()))
		}
		return e, tea.Batch(cmds...)
	case editormsg.FocusMsg:
		cmds = append(cmds, e.Focus())
		switch msg.Model {
		case editormsg.ModelFile:
			f := e.File()
			if f != nil {
				cmds = append(cmds, f.Focus())
			}
			e.searchBar.Blur()
			e.fileTree.Blur()
		case editormsg.ModelSearch:
			cmds = append(cmds, e.searchBar.Focus())
			e.fileTree.Blur()
			f := e.File()
			if f != nil {
				f.Blur()
			}
		case editormsg.ModelFileTree:
			e.fileTree.Focus()
			e.searchBar.Blur()
			f := e.File()
			if f != nil {
				f.Blur()
			}
		}
		return e, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.ToggleFileTree):
			if e.fileTree.Visible() {
				e.fileTree.Hide()
			} else {
				e.fileTree.Show()
				cmds = append(cmds, editormsg.Focus(editormsg.ModelFileTree))
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.FocusFileTree):
			if e.fileTree.Focused() {
				cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
			} else {
				cmds = append(cmds, editormsg.Focus(editormsg.ModelFileTree))
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.File.New):
			return e, overlay.Open(NewNewOverlay())
		case key.Matches(msg, config.Keys.Editor.Search):
			if !e.searchBar.Visible() {
				e.searchBar.Show()
			}
			if !e.searchBar.Focused() {
				cmds = append(cmds, editormsg.Focus(editormsg.ModelSearch))
			} else {
				cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
			}
			return e, tea.Batch(cmds...)
		}
	}

	var cmd tea.Cmd
	e.fileTree, cmd = e.fileTree.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if bubbles.IsKeyMsg(msg) && e.fileTree.Focused() {
		return e, tea.Batch(cmds...)
	}

	e.searchBar, cmd = e.searchBar.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if bubbles.IsKeyMsg(msg) && e.searchBar.Focused() {
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

		if cmd = f.InitTree(); cmd != nil {
			cmds = append(cmds, cmd)
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
	case file.GoToMsg:
		cmds = append(cmds, overlay.Open(NewGoToOverlay(f.Cursor())))
		return e, tea.Batch(cmds...)
	case file.ScrollMsg:
		f.SetCursor(msg.Row, msg.Col)
	case tea.MouseClickMsg:
		for _, z := range append(zone.GetPrefix(file.ZoneFileDiagnosticPrefix), zone.GetPrefix(file.ZoneFileLineDiagnosticPrefix)...) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				index, ok := strings.CutPrefix(z.ID(), file.ZoneFileDiagnosticPrefix)
				if !ok {
					index, _ = strings.CutPrefix(z.ID(), file.ZoneFileLineDiagnosticPrefix)
				}
				i, _ := strconv.Atoi(index)

				diagnostic := f.Diagnostics()[i]
				f.SetCursor(diagnostic.Range.Start.Row, diagnostic.Range.Start.Col)
				f.SetMark(f.Cursor())
				return e, tea.Batch(cmds...)
			}
		}

		for _, z := range zone.GetPrefix(file.ZoneFileLinePrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				row, col := f.GetFileZoneCursorPos(msg, z)
				f.SetCursor(row, col)
				f.SetMark(f.Cursor())
				overwriteCursorBlink = true
			}
		}

		for _, z := range zone.GetPrefix(file.ZoneFileLineNumberPrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), file.ZoneFileLineNumberPrefix))
				f.SetCursor(row, -1)
				f.SetMark(f.Cursor())
				overwriteCursorBlink = true
			}
		}
	case tea.MouseReleaseMsg:
		for _, z := range append(zone.GetPrefix(file.ZoneFileDiagnosticPrefix), zone.GetPrefix(file.ZoneFileLineDiagnosticPrefix)...) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				if !f.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
				}

				index, ok := strings.CutPrefix(z.ID(), file.ZoneFileDiagnosticPrefix)
				if !ok {
					index, _ = strings.CutPrefix(z.ID(), file.ZoneFileLineDiagnosticPrefix)
				}
				i, _ := strconv.Atoi(index)

				if s := f.Selection(); s != nil && !s.Zero() {
					return e, tea.Batch(cmds...)
				}

				diagnostic := f.Diagnostics()[i]
				f.SetCursor(diagnostic.Range.Start.Row, diagnostic.Range.Start.Col)
				f.ShowCurrentDiagnostic()
				return e, tea.Batch(cmds...)
			}
		}

		for _, z := range zone.GetPrefix(file.ZoneFileLinePrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				if !f.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
				}

				row, col := f.GetFileZoneCursorPos(msg, z)
				f.SetCursor(row, col)
				if s := f.Selection(); s == nil || s.Zero() {
					f.ResetMark()
				}
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseRight):
				// TODO: open context menu?
				return e, tea.Batch(cmds...)
			}
		}

		for _, z := range zone.GetPrefix(file.ZoneFileLineNumberPrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				if !f.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
				}

				row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), file.ZoneFileLineNumberPrefix))
				f.SetCursor(row, -1)
				if s := f.Selection(); s == nil || s.Zero() {
					f.ResetMark()
				}
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			}
		}

		for _, z := range zone.GetPrefix(file.ZoneFileLineEmptyPrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				if !f.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
				}
				return e, tea.Batch(cmds...)
			}
		}

		for _, z := range zone.GetPrefix(ZoneFilePrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				if !f.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
				}

				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFilePrefix))
				e.SetFile(i)
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseRight):
				if !f.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
				}

				// TODO: open context menu?
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseMiddle):
				if !f.Focused() {
					cmds = append(cmds, editormsg.Focus(editormsg.ModelFile))
				}

				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFilePrefix))
				cmds = append(cmds, file.CloseFile(e.files[i].Name()))
				return e, tea.Batch(cmds...)
			}
		}

		switch {
		case mouse.Matches(msg, ZoneFileLanguage, tea.MouseLeft):
			log.Println("file language zone")
			cmds = append(cmds, overlay.Open(NewSetLanguageOverlay()))
			return e, tea.Batch(cmds...)
		case mouse.Matches(msg, ZoneFileLineEnding, tea.MouseLeft):
			log.Println("file line ending zone")
			// cmds = append(cmds, overlay.Open(NewSetLineEndingOverlay()))
			return e, tea.Batch(cmds...)
		case mouse.Matches(msg, ZoneFileEncoding, tea.MouseLeft):
			log.Println("file encoding zone")
			// cmds = append(cmds, overlay.Open(NewSetEncodingOverlay()))
			return e, tea.Batch(cmds...)
		case mouse.Matches(msg, ZoneFileGoTo, tea.MouseLeft):
			cmds = append(cmds, overlay.Open(NewGoToOverlay(f.Cursor())))
			return e, tea.Batch(cmds...)
		}
	case tea.MouseMotionMsg:
		for _, z := range zone.GetPrefix(file.ZoneFileLinePrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				row, col := f.GetFileZoneCursorPos(msg, z)
				f.SetCursor(row, col)
				return e, tea.Batch(cmds...)
			}
		}
	case tea.MouseWheelMsg:
		for _, z := range append(zone.GetPrefix(file.ZoneFileLinePrefix), zone.GetPrefix(file.ZoneFileLineNumberPrefix)...) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseWheelLeft), mouse.MatchesZone(msg, z, tea.MouseWheelDown, tea.ModShift):
				f.MoveCursorLeft(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseWheelRight), mouse.MatchesZone(msg, z, tea.MouseWheelUp, tea.ModShift):
				f.MoveCursorRight(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseWheelUp):
				f.MoveCursorUp(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseWheelDown):
				f.MoveCursorDown(1)
				cmds = append(cmds, f.Autocomplete().Update())
				return e, tea.Batch(cmds...)
			}
		}
	case tea.KeyPressMsg:
		if e.focus {
			switch {
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Show):
				row, col := f.Cursor()
				cmds = append(cmds, ls.GetAutocompletion(f.Name(), row, col))
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Cancel) && f.Autocomplete().Visible():
				f.Autocomplete().ClearCompletions()
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Next) && f.Autocomplete().Visible():
				f.Autocomplete().Next()
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Prev) && f.Autocomplete().Visible():
				f.Autocomplete().Previous()
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Apply) && f.Autocomplete().Visible():
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
				if cmd = f.InitTree(); cmd != nil {
					cmds = append(cmds, cmd)
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
			case key.Matches(msg, config.Keys.Editor.Diagnostic.Show):
				f.ShowCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Cancel) && f.ShowsCurrentDiagnostic():
				f.HideCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Editor.Code.ShowDeclaration):
				cmds = append(cmds, f.ShowDeclaration())
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowDefinitions):
				cmds = append(cmds, f.ShowDefinitions())
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowTypeDefinition):
				cmds = append(cmds, f.ShowTypeDefinitions())
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowImplementation):
				// cmds = append(cmds, f.ShowImplementations())
				// return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowReferences):
				// cmds = append(cmds, f.ShowReferences())
				// return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.OpenOutline):
				cmds = append(cmds, overlay.Open(NewOutlineOverlay(f)))
			case key.Matches(msg, config.Keys.Editor.File.Next):
				if e.activeFile < len(e.files)-1 {
					e.activeFile++
					f.Blur()
					cmds = append(cmds, e.files[e.activeFile].Focus())
				}
			case key.Matches(msg, config.Keys.Editor.File.Prev):
				if e.activeFile > 0 {
					e.activeFile--
					f.Blur()
					cmds = append(cmds, e.files[e.activeFile].Focus())
				}
			case key.Matches(msg, config.Keys.Editor.File.Close):
				cmds = append(cmds, file.Close)

			case key.Matches(msg, config.Keys.Editor.File.Delete):
				cmds = append(cmds, overlay.Open(NewDeleteOverlay()))
			case key.Matches(msg, config.Keys.Editor.File.Rename):
				cmds = append(cmds, overlay.Open(NewRenameOverlay(f.Name())))
			case key.Matches(msg, config.Keys.Editor.Navigation.LineUp):
				f.MoveCursorUp(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.LineDown):
				f.MoveCursorDown(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.CharacterLeft):
				f.MoveCursorLeft(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.CharacterRight):
				f.MoveCursorRight(moveSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.WordUp):
				f.MoveCursorWordUp()
			case key.Matches(msg, config.Keys.Editor.Navigation.WordDown):
				f.MoveCursorWordDown()
			case key.Matches(msg, config.Keys.Editor.Navigation.WordLeft):
				f.SetCursor(f.NextWordLeft())
			case key.Matches(msg, config.Keys.Editor.Navigation.WordRight):
				f.SetCursor(f.NextWordRight())
			case key.Matches(msg, config.Keys.Editor.Navigation.PageUp):
				f.MoveCursorUp(pageSize)
			case key.Matches(msg, config.Keys.Editor.Navigation.PageDown):
				f.MoveCursorDown(pageSize)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.LineStart):
				f.SetCursor(-1, 0)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.LineEnd):
				cursorRow, _ := f.Cursor()
				f.SetCursor(-1, f.Buffer().LineLen(cursorRow))
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.FileStart):
				f.SetCursor(0, -1)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.FileEnd):
				f.SetCursor(f.Buffer().LinesLen(), -1)
				cmds = append(cmds, f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Navigation.GoTo):
				cmds = append(cmds, overlay.Open(NewGoToOverlay(f.Cursor())))
				return e, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Edit.Copy):
				selBytes := f.SelectionBytes()
				if len(selBytes) > 0 {
					cmds = append(cmds, file.Copy(selBytes))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.Paste):
				cmds = append(cmds, file.Paste)
			case key.Matches(msg, config.Keys.Editor.Edit.Cut):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, file.Cut(*s, f.SelectionBytes()))
				}
			case key.Matches(msg, config.Keys.Editor.Selection.SelectLeft):
				f.SelectLeft(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectRight):
				f.SelectRight(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectUp):
				f.SelectUp(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectDown):
				f.SelectDown(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectAll):
				f.SelectAll()

			case key.Matches(msg, config.Keys.Editor.File.Save):
				cmds = append(cmds, file.SaveFile(f.Name()))
			case key.Matches(msg, config.Keys.Editor.Edit.Tab):
				cmds = append(cmds, f.InsertRunes([]rune{'\t'}))
			case key.Matches(msg, config.Keys.Editor.Edit.RemoveTab):
				cmds = append(cmds, f.RemoveTab())
			case key.Matches(msg, config.Keys.Editor.Edit.Newline):
				f.ResetMark()
				cmds = append(cmds, f.InsertNewLine(), f.Autocomplete().Update())
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteRight):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteAfter(1))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteLeft):
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
							closeWidth := ansi.StringWidth(pair.Close)
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
			case key.Matches(msg, config.Keys.Editor.Edit.DuplicateLine):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.Insert(f.SelectionBytes()))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DuplicateLine())
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteWordLeft):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteWordLeft())
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteWordRight):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteWordRight())
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteLine):
				s := f.Selection()
				if s != nil {
					cmds = append(cmds, f.DeleteRange(s.Start, s.End))
					f.ResetMark()
				} else {
					cmds = append(cmds, f.DeleteLine())
				}
			case key.Matches(msg, config.Keys.Editor.Edit.ToggleComment):
				cmds = append(cmds, f.ToggleComment())
				overwriteCursorBlink = true
			case key.Matches(msg, config.Keys.Debug):
				log.Println("DEBUG")

			default:
				k := msg.Key()
				if k.Text == "" || k.Mod.Contains(tea.ModAlt) || k.Mod.Contains(tea.ModCtrl) || k.Mod.Contains(tea.ModMeta) || k.Mod.Contains(tea.ModSuper) || k.Mod.Contains(tea.ModHyper) {
					break
				}

				text := []byte(k.Text)
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
						if string(k.Code) == pair.Open {
							row, col := f.Cursor()
							cmds = append(cmds, f.InsertAt(row, col+ansi.StringWidth(pair.Open), []byte(pair.Close)))
							break
						}
					}
				}
			}
		}
	}

	if cmd = f.UpdateCursor(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	newRow, newCol := f.Cursor()
	if oldRow != newRow || oldCol != newCol || overwriteCursorBlink {
		f.SetCursorBlink(false)
		cmds = append(cmds, f.CursorBlinkCmd())
	}

	return e, tea.Batch(cmds...)
}

func (e *Editor) View(width int, height int) string {
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
			Render(fmt.Sprintf("No file open.\n\nPress '%s' to open a file.", config.Keys.Editor.File.Open.Help().Key))

		if fileTree == "" {
			return code
		}
		code = config.Theme.UI.FileView.BorderStyle.Render(code)
		return lipgloss.JoinHorizontal(lipgloss.Top, fileTree, code)
	}

	var searchBar string
	if e.searchBar.Visible() {
		searchBar = e.searchBar.View()
		if fileTree != "" {
			searchBar = config.Theme.UI.FileView.BorderStyle.Render(searchBar)
		}
		height -= lipgloss.Height(searchBar)
	}

	editor := f.View(width, height, e.fileTree.Visible(), e.treeSitterDebug)

	if searchBar != "" {
		editor = lipgloss.JoinVertical(lipgloss.Left, searchBar, editor)
	}
	if fileTree != "" {
		return lipgloss.JoinHorizontal(lipgloss.Top, fileTree, editor)
	}

	return editor
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

		style := config.Theme.UI.AppBar.Files.FileStyle
		if i == e.activeFile {
			style = config.Theme.UI.AppBar.Files.SelectedFileStyle
		}

		fileName := clampString(f.FileName(), 16)
		if f.Dirty() {
			fileName += "*"
		} else {
			fileName += " "
		}

		fileName = fmt.Sprintf("%s%s", icon, style.Inline(true).Render(" "+fileName))

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
