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
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lrstanley/bubblezone"

	"go.gopad.dev/gopad/gopad/editor/buffer"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/doc"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles"
	"go.gopad.dev/gopad/internal/bubbles/key"
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
		searchBar: newSearchBar(),
		fileTree:  newFileTree(),
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
	workspace string
	args      []string

	fileTree  fileTree
	searchBar searchBar

	docViews  []*DocumentView
	activeDoc int
	docOffset int

	focus ModelType
	debug bool
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

		arg, _ = filepath.Abs(arg)

		cmd, err := e.OpenFile(arg)
		if err != nil {
			return e, notifications.Addf("error while opening file %s: %s", arg, err)
		}
		cmds = append(cmds, cmd)
	}

	if f := e.DocView(); f != nil {
		cmds = append(cmds, Focus(ModelTypeFile))
	} else {
		cmds = append(cmds, Focus(ModelTypeFileTree))
	}

	return e, tea.Batch(cmds...)
}

func (e Editor) Workspace() string {
	return e.workspace
}

func (e *Editor) Focused() bool {
	return e.focus != ModelTypeNone
}

func (e *Editor) Focus(model ModelType) tea.Cmd {
	e.focus = model

	var cmds []tea.Cmd
	switch model {
	case ModelTypeFile:
		e.fileTree.Blur()
		e.searchBar.Blur()
		if f := e.DocView(); f != nil {
			cmds = append(cmds, f.Focus())
		}
	case ModelTypeFileTree:
		e.fileTree.Focus()
		e.searchBar.Blur()
		if f := e.DocView(); f != nil {
			cmds = append(cmds, f.Blur())
		}
	case ModelTypeSearchBar:
		e.fileTree.Blur()
		cmds = append(cmds, e.searchBar.Focus())
		if f := e.DocView(); f != nil {
			cmds = append(cmds, f.Blur())
		}
	default:
		panic(fmt.Sprintf("unknown model type: %d", model))
	}

	return tea.Batch(cmds...)
}

func (e *Editor) Blur() tea.Cmd {
	var cmds []tea.Cmd

	e.focus = ModelTypeNone

	e.fileTree.Blur()
	e.searchBar.Blur()

	f := e.DocView()
	if f != nil {
		cmds = append(cmds, f.Blur())
	}

	return tea.Batch(cmds...)
}

func (e *Editor) CreateFile(name string) (tea.Cmd, error) {
	if !filepath.IsAbs(name) {
		name = filepath.Join(e.workspace, name)
		name, _ = filepath.Abs(name)
	}
	if slices.ContainsFunc(e.docViews, func(b *DocumentView) bool {
		return b.Doc.Name == name
	}) {
		return nil, nil
	}

	buff, err := buffer.New(bytes.NewReader(nil), buffer.LineEndingLF)
	if err != nil {
		return nil, err
	}
	v, err := newDocumentView(name, buff, doc.ModeWrite)
	if err != nil {
		return nil, err
	}

	e.docViews = append(e.docViews, v)

	cmds := []tea.Cmd{
		tea.Sequence(
			ls.FileCreated(v.Doc.Name, v.Doc.Buffer.Bytes()),
			ls.FileOpened(v.Doc.Name, v.Doc.Version(), v.LanguageName(), v.Doc.Buffer.Bytes()),
			ls.GetInlayHint(v.Doc.Name, v.Doc.Version(), v.Doc.Range()),
		),
	}

	return tea.Batch(cmds...), nil
}

func (e *Editor) OpenFile(name string) (tea.Cmd, error) {
	if slices.ContainsFunc(e.docViews, func(b *DocumentView) bool {
		return b.Doc.Name == name
	}) {
		return nil, nil
	}

	v, err := newDocumentViewFromName(name)
	if err != nil {
		return nil, err
	}
	e.docViews = append(e.docViews, v)

	cmds := []tea.Cmd{
		tea.Sequence(
			ls.FileOpened(v.Doc.Name, v.Doc.Version(), v.LanguageName(), v.Doc.Buffer.Bytes()),
			ls.GetInlayHint(v.Doc.Name, v.Doc.Version(), v.Doc.Range()),
		),
	}

	return tea.Batch(cmds...), nil
}

func (e *Editor) FormatFile(name string) (tea.Cmd, error) {
	f := e.DocViewByName(name)
	if f == nil {
		return nil, nil
	}
	cmd, err := f.Doc.Format()
	if err != nil {
		return cmd, err
	}

	return tea.Batch(cmd, ls.FileCreated(f.Doc.Name, f.Doc.Buffer.Bytes())), nil
}

func (e *Editor) SaveFile(name string) (tea.Cmd, error) {
	f := e.DocViewByName(name)
	if f == nil {
		return nil, nil
	}
	if err := f.Doc.Save(); err != nil {
		return nil, err
	}

	return ls.FileSaved(f.Doc.Name, f.Doc.Buffer.Bytes()), nil
}

func (e *Editor) RenameFile(oldName string, newName string) (tea.Cmd, error) {
	if !filepath.IsAbs(newName) {
		newName = filepath.Join(e.workspace, newName)
		newName, _ = filepath.Abs(newName)
	}
	f := e.DocViewByName(oldName)
	if f == nil {
		return nil, nil
	}
	if err := f.Doc.Rename(newName); err != nil {
		return nil, err
	}

	return ls.FileRenamed(oldName, newName), nil
}

func (e *Editor) CloseFile(name string) (tea.Cmd, error) {
	index := slices.IndexFunc(e.docViews, func(file *DocumentView) bool {
		return file.Doc.Name == name
	})
	if index == -1 {
		return nil, nil
	}

	var cmds []tea.Cmd
	f := e.docViews[index]
	e.docViews = slices.Delete(e.docViews, index, index+1)
	e.activeDoc = min(e.activeDoc, len(e.docViews)-1)
	if len(e.docViews) > 0 {
		cmds = append(cmds, e.docViews[e.activeDoc].Focus())
	} else {
		e.fileTree.Focus()
	}

	return tea.Batch(
		tea.Sequence(cmds...),
		ls.FileClosed(f.Doc.Name),
	), nil
}

func (e *Editor) DeleteFile(name string) (tea.Cmd, error) {
	index := slices.IndexFunc(e.docViews, func(file *DocumentView) bool {
		return file.Doc.Name == name
	})
	if index == -1 {
		return nil, nil
	}

	f := e.docViews[index]
	if err := f.Doc.Delete(); err != nil {
		return nil, err
	}

	var cmds []tea.Cmd
	e.docViews = slices.Delete(e.docViews, e.activeDoc, e.activeDoc+1)
	e.activeDoc = min(e.activeDoc, len(e.docViews)-1)
	if len(e.docViews) > 0 {
		cmds = append(cmds, e.docViews[e.activeDoc].Focus())
	} else {
		e.fileTree.Focus()
	}

	return tea.Batch(
		tea.Sequence(cmds...),
		ls.FileDeleted(f.Doc.Name),
	), nil
}

func (e *Editor) DocView() *DocumentView {
	if len(e.docViews) == 0 {
		return nil
	}
	return e.docViews[e.activeDoc]
}

func (e *Editor) SetDocView(index int) tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, e.docViews[e.activeDoc].Blur())
	e.activeDoc = index
	cmds = append(cmds, e.docViews[index].Focus())
	e.docViews[index].resetCursor()
	return tea.Sequence(cmds...)
}

func (e *Editor) SetDocViewByName(name string) tea.Cmd {
	for i, f := range e.docViews {
		if f.Doc.Name == name {
			return e.SetDocView(i)
		}
	}

	return nil
}

func (e *Editor) DocViewByName(name string) *DocumentView {
	for _, f := range e.docViews {
		if f.Doc.Name == name {
			return f
		}
	}
	return nil
}

func (e *Editor) HasChanges() bool {
	for _, f := range e.docViews {
		if f.Doc.Buffer.Dirty() {
			return true
		}
	}
	return false
}

func (e *Editor) ToggleDebug() {
	e.debug = !e.debug
}

func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case SetLanguageActionMsg:
		if v := e.DocView(); v != nil {
			cmds = append(cmds, v.SetLanguage(msg.Language))
		}
		return e, tea.Batch(cmds...)
	case FormatActionMsg:
		if v := e.DocView(); v != nil {
			cmds = append(cmds, FormatFile(v.Doc.Name))
		}
		return e, tea.Batch(cmds...)
	case SaveActionMsg:
		if v := e.DocView(); v != nil {
			cmds = append(cmds, SaveFile(v.Doc.Name))
		}
		return e, tea.Batch(cmds...)
	case RenameActionMsg:
		if v := e.DocView(); v != nil {
			cmds = append(cmds, overlay.Open(NewRenameOverlay(v.Doc.Name)))
		}
		return e, tea.Batch(cmds...)
	case DeleteActionMsg:
		if v := e.DocView(); v != nil {
			if v.Doc.Buffer.Dirty() {
				return e, overlay.Open(NewDeleteOverlay([]string{v.Doc.Name}))
			}
			cmds = append(cmds, CloseFile(v.Doc.Name))
		}
		return e, tea.Batch(cmds...)
	case CloseActionMsg:
		if v := e.DocView(); v != nil {
			if v.Doc.Buffer.Dirty() {
				return e, overlay.Open(NewCloseOverlay([]string{v.Doc.Name}))
			}
			cmds = append(cmds, CloseFile(v.Doc.Name))
		}
		return e, tea.Batch(cmds...)
	case GoToActionMsg:
		if v := e.DocView(); v != nil {
			cmds = append(cmds, overlay.Open(NewGoToOverlay(v.Doc.Cursor())))
		}
		return e, tea.Batch(cmds...)
	case SelectActionMsg:
		if v := e.DocView(); v != nil {
			v.Doc.SetMark(msg.Start)
			v.Doc.SetCursor(msg.End)
		}
		return e, tea.Batch(cmds...)
	case ScrollActionMsg:
		if v := e.DocView(); v != nil {
			v.Doc.SetCursor(buffer.Point(msg))
		}
		return e, tea.Batch(cmds...)

	case CutMsg:
		if v := e.DocView(); v != nil {
			v.Doc.DeleteRange(buffer.Range(msg))
			v.Doc.ResetMark()
		}
		return e, tea.Batch(cmds...)
	case tea.PasteMsg:
		if v := e.DocView(); v != nil {
			s := v.Doc.Selection()
			if s != nil {
				v.Doc.Replace(*s, []byte(msg))
				v.Doc.ResetMark()
			} else {
				v.Doc.Insert(v.Doc.Cursor(), []byte(msg))
			}
		}
		return e, tea.Batch(cmds...)

	case OpenDirMsg:
		e.fileTree.Show()
		e.fileTree.Focus()

		if f := e.DocView(); f != nil {
			cmds = append(cmds, f.Blur())
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
	case OpenFileMsg:
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
			Focus(ModelTypeFile),
		)
		cmds = append(cmds, e.SetDocViewByName(msg.Name))
		if msg.Position != nil {
			e.DocView().Doc.SetCursor(*msg.Position)
		}
		return e, tea.Batch(cmds...)
	case FormatFileMsg:
		cmd, err := e.FormatFile(msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while formatting file %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s formatted", msg.Name)))
		return e, tea.Batch(cmds...)
	case SaveFileMsg:
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
	case NewFileMsg:
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
			Focus(ModelTypeFile),
		)
		cmds = append(cmds, e.SetDocViewByName(msg.Name))
		return e, tea.Batch(cmds...)
	case RenameFileMsg:
		cmd, err := e.RenameFile(msg.OldName, msg.NewName)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while renamed file %s: %s", msg.OldName, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s renamed to %s", msg.OldName, msg.NewName)))
		return e, tea.Batch(cmds...)
	case DeleteFileMsg:
		cmd, err := e.DeleteFile(msg.Name)
		if err != nil {
			cmds = append(cmds, notifications.Add(fmt.Sprintf("error while deleting file %s: %s", msg.Name, err.Error())))
			return e, tea.Batch(cmds...)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, notifications.Add(fmt.Sprintf("file %s deleted", msg.Name)))
	case CloseFileMsg:
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
	case CloseAllActionMsg:
		var files []string
		for _, f := range e.docViews {
			if f.Doc.Buffer.Dirty() {
				files = append(files, f.Doc.Name)
			}
		}
		if len(files) > 0 {
			return e, overlay.Open(NewCloseOverlay(files))
		}
		fileCmds := make([]tea.Cmd, len(e.docViews))
		for _, f := range e.docViews {
			fileCmds = append(fileCmds, CloseFile(f.Doc.Name))
		}
		cmds = append(cmds, tea.Sequence(fileCmds...))
		return e, tea.Batch(cmds...)

	case FocusMsg:
		cmds = append(cmds, e.Focus(msg.Model))
		return e, tea.Batch(cmds...)

	case tea.MouseReleaseMsg:
		for _, z := range zone.GetPrefix(ZoneFilePrefix) {
			switch {
			case mouse.MatchesZone(msg, z, tea.MouseLeft):
				cmds = append(cmds, Focus(ModelTypeFile))

				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFilePrefix))
				cmds = append(cmds, e.SetDocView(i))
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseRight):
				// TODO: open context menu?
				log.Println("right click on file")
				return e, tea.Batch(cmds...)
			case mouse.MatchesZone(msg, z, tea.MouseMiddle):
				i, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFilePrefix))
				cmds = append(cmds, CloseFile(e.docViews[i].Doc.Name))
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
			f := e.DocView()
			if f == nil {
				return e, tea.Batch(cmds...)
			}
			cmds = append(cmds, overlay.Open(NewGoToOverlay(f.Doc.Cursor())))
			return e, tea.Batch(cmds...)
		}

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, config.Keys.Editor.ToggleFileTree):
			if e.fileTree.Visible() {
				e.fileTree.Hide()
			} else {
				e.fileTree.Show()
				cmds = append(cmds, Focus(ModelTypeFileTree))
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.FocusFileTree):
			if e.fileTree.Focused() {
				cmds = append(cmds, Focus(ModelTypeFile))
			} else {
				cmds = append(cmds, Focus(ModelTypeFileTree))
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.File.New):
			return e, overlay.Open(NewNewOverlay())
		case key.Matches(msg, config.Keys.Editor.Search):
			if !e.searchBar.Visible() {
				e.searchBar.Show()
			}
			if !e.searchBar.Focused() {
				cmds = append(cmds, Focus(ModelTypeSearchBar))
			} else {
				cmds = append(cmds, Focus(ModelTypeFile))
			}
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.ToggleDebug):
			e.ToggleDebug()
			return e, tea.Batch(cmds...)
		case key.Matches(msg, config.Keys.Editor.DebugTreeSitterNodes):
			v := e.DocView()
			if v != nil {
				if v.Doc.Syntax == nil {
					cmds = append(cmds, notifications.Add("no syntax tree available for this file"))
					return e, tea.Batch(cmds...)
				}
				tree := v.Doc.Syntax.Layers.Tree()

				buff, err := buffer.New(bytes.NewReader([]byte(tree.RootNode().ToSexp())), buffer.LineEndingLF)
				if err != nil {
					cmds = append(cmds, notifications.Addf("error while creating tree s-expression buffer: %s", err.Error()))
					return e, tea.Batch(cmds...)
				}

				dv, err := newDocumentView(v.Doc.FileName()+".sexp", buff, doc.ModeReadOnly)
				if err != nil {
					cmds = append(cmds, notifications.Addf("error while creating tree s-expression view: %s", err.Error()))
					return e, tea.Batch(cmds...)
				}

				e.docViews = append(e.docViews, dv)
				e.activeDoc = len(e.docViews) - 1
			}
		case key.Matches(msg, config.Keys.Editor.File.Next):
			if e.activeDoc < len(e.docViews)-1 {
				cmds = append(cmds, e.SetDocView(e.activeDoc+1))
			}
		case key.Matches(msg, config.Keys.Editor.File.Prev):
			if e.activeDoc > 0 {
				cmds = append(cmds, e.SetDocView(e.activeDoc-1))
			}
		}
	}

	var cmd tea.Cmd
	if e.focus == ModelTypeFileTree || !bubbles.IsKeyMsg(msg) {
		e.fileTree, cmd = e.fileTree.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if e.focus == ModelTypeSearchBar || !bubbles.IsKeyMsg(msg) {
		e.searchBar, cmd = e.searchBar.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if e.focus == ModelTypeFile || !bubbles.IsKeyMsg(msg) {
		for i, fileView := range e.docViews {
			if i != e.activeDoc && bubbles.IsInputMsg(msg) {
				continue
			}

			updatedFileView, cmd := fileView.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			e.docViews[i] = &updatedFileView
		}
	}
	return e, tea.Batch(cmds...)
}

func (e *Editor) View(width int, height int, offsetX int, offsetY int) string {
	var tree string
	if e.fileTree.Visible() {
		tree = e.fileTree.View(height)
		treeWidth := lipgloss.Width(tree)
		width -= treeWidth
		offsetX += treeWidth
	}

	f := e.DocView()
	if f == nil {
		width -= config.Theme.UI.FileView.EmptyStyle.GetHorizontalBorderSize()
		height -= config.Theme.UI.FileView.EmptyStyle.GetVerticalBorderSize()

		code := config.Theme.UI.FileView.EmptyStyle.
			Width(width).
			Height(height).
			Render(fmt.Sprintf("No file open.\n\nPress '%s' to open a file.", config.Keys.Editor.File.Open.Help().Key))

		if tree == "" {
			return code
		}
		code = config.Theme.UI.FileView.BorderStyle.Render(code)
		return lipgloss.JoinHorizontal(lipgloss.Top, tree, code)
	}

	var search string
	if e.searchBar.Visible() {
		search = e.searchBar.View()
		if tree != "" {
			search = config.Theme.UI.FileView.BorderStyle.Render(search)
		}
		searchHeight := lipgloss.Height(search)
		height -= searchHeight
		offsetY += searchHeight
	}

	editor := f.View(width, height, e.fileTree.Visible(), e.debug, offsetX, offsetY)

	if search != "" {
		editor = lipgloss.JoinVertical(lipgloss.Left, search, editor)
	}
	if tree != "" {
		return lipgloss.JoinHorizontal(lipgloss.Top, tree, editor)
	}

	return editor
}

func (e *Editor) refreshActiveFileOffset(width int, fileNames []string) {
	if e.docOffset == 0 {
		return
	}

	filesWidth := lipgloss.Width(strings.Join(fileNames[:e.activeDoc], ""))
	if filesWidth < width {
		e.docOffset = 0
	}

	if filesWidth > width {
		e.docOffset = e.activeDoc
	}

	for i := e.docOffset; i > 0; i-- {
		filesWidth = lipgloss.Width(strings.Join(fileNames[i:e.activeDoc], ""))
		if filesWidth < width {
			e.docOffset = i
			break
		}
	}
}

func (e *Editor) FileTabsView(width int) string {
	var fileNames []string
	for i, f := range e.docViews {
		languageName := f.LanguageName()
		icon := config.Theme.Icons.FileIcon(languageName).Render()

		style := config.Theme.UI.AppBar.Files.FileStyle
		if i == e.activeDoc {
			style = config.Theme.UI.AppBar.Files.SelectedFileStyle
		}

		fileName := clampString(f.Doc.FileName(), 16)
		if f.Doc.Buffer.Dirty() {
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

	return config.Theme.UI.AppBar.Files.Style.Render(fileNames[e.docOffset:]...)
}

func clampString(s string, length int) string {
	if len(s) > length {
		return s[:length-1] + "…"
	}
	return s
}
