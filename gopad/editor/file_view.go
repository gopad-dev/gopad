package editor

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/bubblezone"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/file"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/key"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/buffer"
)

const (
	ZoneFileLineEmptyPrefix      = "file.line.empty:"
	ZoneFileLinePrefix           = "file.line:"
	ZoneFileLineNumberPrefix     = "file.line.number:"
	ZoneFileDiagnosticPrefix     = "file.diagnostic:"
	ZoneFileLineDiagnosticPrefix = "file.line.diagnostic:"
)

func zoneFileLineEmptyID(line int) string {
	return fmt.Sprintf("%s%s", ZoneFileLineEmptyPrefix, strconv.Itoa(line))
}

func zoneFileLineID(line int) string {
	return fmt.Sprintf("%s%s", ZoneFileLinePrefix, strconv.Itoa(line))
}

func zoneFileLineNumberID(line int) string {
	return fmt.Sprintf("%s%s", ZoneFileLineNumberPrefix, strconv.Itoa(line))
}

func zoneFileDiagnosticID(id int) string {
	return fmt.Sprintf("%s%s", ZoneFileDiagnosticPrefix, strconv.Itoa(id))
}

func zoneFileLineDiagnosticID(id int) string {
	return fmt.Sprintf("%s%s", ZoneFileLineDiagnosticPrefix, strconv.Itoa(id))
}

func newFileView(buff buffer.Buffer, mode file.Mode) *FileView {
	return &FileView{
		file: file.NewFileWithBuffer(buff, mode),
		cursor: fileCursor{
			point: buffer.Point{
				Row: 0,
				Col: 0,
			},
			cursor: config.NewCursor(),
		},
	}
}

func newFileViewFromName(name string) (*FileView, error) {
	f, err := file.NewFileFromName(name)
	if err != nil {
		return nil, err
	}
	return &FileView{
		file: f,
		cursor: fileCursor{
			point: buffer.Point{
				Row: 0,
				Col: 0,
			},
			cursor: config.NewCursor(),
		},
	}, nil
}

type FileView struct {
	file                  *file.File
	cursor                fileCursor
	showCurrentDiagnostic bool
	definitionsIndex      int
}

func (v *FileView) Name() string {
	return v.file.Buffer.Name()
}

func (v *FileView) RelativeName(workspace string) string {
	return v.file.RelativeName(workspace)
}

func (v *FileView) Language() *file.Language {
	return v.file.Language
}

func (v *FileView) LineEnding() buffer.LineEnding {
	return v.file.Buffer.LineEnding()
}

func (v *FileView) EncodingName() string {
	return v.file.Buffer.EncodingName()
}

func (v *FileView) Focus() tea.Cmd {
	return v.cursor.cursor.Focus()
}

func (v *FileView) Blur() {
	v.cursor.cursor.Blur()
}

func (v FileView) Focused() bool {
	return v.cursor.cursor.Focused()
}

func (v *FileView) ShowsCurrentDiagnostic() bool {
	return v.showCurrentDiagnostic
}

func (v *FileView) ShowCurrentDiagnostic() {
	v.showCurrentDiagnostic = true
}

func (v *FileView) HideCurrentDiagnostic() {
	v.showCurrentDiagnostic = false
}

func (v FileView) GetCursorForCharPos(p buffer.Point) buffer.Point {
	positionRow := max(p.Row-v.cursor.offset.Row, 0)
	if positionRow >= len(v.file.Positions) {
		return buffer.Point{
			Row: max(v.file.Buffer.LinesLen()-1, 0),
			Col: 0,
		}
	}

	linePositions := v.file.Positions[positionRow]
	if p.Col >= len(linePositions) {
		return buffer.Point{
			Row: p.Row,
			Col: v.file.Buffer.LineLen(p.Row),
		}
	}

	return linePositions[p.Col]
}

func (v FileView) GetFileZoneCursorPos(msg tea.MouseMsg, z *zone.ZoneInfo) buffer.Point {
	row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFileLinePrefix))
	col, _ := z.Pos(msg)
	return v.GetCursorForCharPos(buffer.Point{Row: row, Col: col})
}

func (v *FileView) SetLanguage(language string) tea.Cmd {
	v.file.SetLanguage(language)
	v.file.ClearDiagnosticsByType(ls.DiagnosticTypeTreeSitter)

	return v.file.InitTree()
}

func (v *FileView) refreshCursorViewOffset(width int, height int) {
	c := v.Cursor()

	// TODO: figure out how to handle inlay hints when scrolling horizontally
	// if len(f.positions) > 0 {
	//	i, ok := slices.BinarySearchFunc(f.positions[cursorRow], cursorCol, func(p pos, col int) int {
	//		if p.col < col {
	//			return -1
	//		} else if p.col > col {
	//			return 1
	//		}
	//		return 0
	//	})
	//	if ok {
	//		cursorCol = i
	//	}
	// }

	if c.Row >= v.cursor.offset.Row+height {
		v.cursor.offset.Row = c.Row - height + 1
	} else if c.Row < v.cursor.offset.Row {
		v.cursor.offset.Row = c.Row
	}

	if c.Col >= v.cursor.offset.Col+width {
		v.cursor.offset.Col = c.Col - width + 1
	} else if c.Col < v.cursor.offset.Col {
		v.cursor.offset.Col = c.Col
	}
}

func (v FileView) Update(msg tea.Msg) (FileView, tea.Cmd) {
	var cmds []tea.Cmd
	var overwriteCursorBlink bool

	oldCursor := v.Cursor()

	switch msg := msg.(type) {
	case ls.UpdateFileDiagnosticMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}
		v.file.SetDiagnostic(msg.Type, msg.Version, msg.Diagnostics)
		return v, tea.Batch(cmds...)
	case ls.UpdateAutocompletionMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}
		v.file.Autocomplete.SetCompletions(msg.Completions)
		return v, tea.Batch(cmds...)
	case ls.UpdateInlayHintMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}
		v.file.SetInlayHint(msg.Version, msg.Hints)
		return v, tea.Batch(cmds...)
	case file.UpdateMatchesMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}
		v.file.SetMatches(msg.Version, msg.Matches)
		return v, tea.Batch(cmds...)
	case ls.RefreshInlayHintMsg:
		cmds = append(cmds, ls.GetInlayHint(v.Name(), v.file.Buffer.Version(), v.file.Range()))
		return v, tea.Batch(cmds...)
	case ls.UpdateDeclarationsMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}
		if len(msg.Declarations) == 0 {
			cmds = append(cmds, notifications.Add("No declaration found"))
		} else if len(msg.Declarations) == 1 {
			declaration := msg.Declarations[0]
			cmds = append(cmds, OpenFilePosition(declaration.Name, &buffer.Point{
				Row: declaration.Range.Start.Row,
				Col: declaration.Range.Start.Col,
			}), notifications.Add("Found declaration"))

			return v, tea.Batch(cmds...)
		}

		v.file.SetDeclarations(msg.Declarations)
		return v, tea.Batch(cmds...)
	case ls.UpdateDefinitionsMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}

		if len(msg.Definitions) == 0 {
			cmds = append(cmds, notifications.Add("No definition found"))
		} else if len(msg.Definitions) == 1 {
			definition := msg.Definitions[0]
			cmds = append(cmds, OpenFilePosition(definition.Name, &buffer.Point{
				Row: definition.Range.Start.Row,
				Col: definition.Range.Start.Col,
			}), notifications.Add("Found definition"))

			return v, tea.Batch(cmds...)
		}
		v.file.SetDefinitions(msg.Definitions)
		return v, tea.Batch(cmds...)
	case ls.UpdateTypeDefinitionsMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}

		if len(msg.TypeDefinitions) == 0 {
			cmds = append(cmds, notifications.Add("No type definition found"))
		} else if len(msg.TypeDefinitions) == 1 {
			typeDefinition := msg.TypeDefinitions[0]
			cmds = append(cmds, OpenFilePosition(typeDefinition.Name, &buffer.Point{
				Row: typeDefinition.Range.Start.Row,
				Col: typeDefinition.Range.Start.Col,
			}), notifications.Add("Found type definition"))

			return v, tea.Batch(cmds...)
		}

		v.file.SetTypeDefinitions(msg.TypeDefinitions)
		return v, tea.Batch(cmds...)
	case ls.UpdateImplementationsMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}

		if len(msg.Implementations) == 0 {
			cmds = append(cmds, notifications.Add("No implementation found"))
		} else if len(msg.Implementations) == 1 {
			implementation := msg.Implementations[0]
			cmds = append(cmds, OpenFilePosition(implementation.Name, &buffer.Point{
				Row: implementation.Range.Start.Row,
				Col: implementation.Range.Start.Col,
			}), notifications.Add("Found implementation"))

			return v, tea.Batch(cmds...)
		}

		v.file.SetImplementations(msg.Implementations)
		return v, tea.Batch(cmds...)
	case ls.UpdateReferencesMsg:
		if msg.Name != v.Name() {
			return v, tea.Batch(cmds...)
		}

		if len(msg.References) == 0 {
			cmds = append(cmds, notifications.Add("No references found"))
		} else if len(msg.References) == 1 {
			reference := msg.References[0]
			cmds = append(cmds, OpenFilePosition(reference.Name, &buffer.Point{
				Row: reference.Range.Start.Row,
				Col: reference.Range.Start.Col,
			}), notifications.Add("Found reference"))

			return v, tea.Batch(cmds...)
		}

		v.file.SetReferences(msg.References)
		return v, tea.Batch(cmds...)

	case tea.MouseMsg:
		switch msg := msg.(type) {
		case tea.MouseClickMsg:
			for _, z := range append(zone.GetPrefix(ZoneFileDiagnosticPrefix), zone.GetPrefix(ZoneFileLineDiagnosticPrefix)...) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					index, ok := strings.CutPrefix(z.ID(), ZoneFileDiagnosticPrefix)
					if !ok {
						index, _ = strings.CutPrefix(z.ID(), ZoneFileLineDiagnosticPrefix)
					}
					i, _ := strconv.Atoi(index)

					diagnostic := v.file.Diagnostics[i]
					v.SetCursor(diagnostic.Range.Start)
					v.SetMark(v.Cursor())
					return v, tea.Batch(cmds...)
				}
			}

			for _, z := range zone.GetPrefix(ZoneFileLinePrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					p := v.GetFileZoneCursorPos(msg, z)
					v.SetCursor(p)
					v.SetMark(v.Cursor())
					overwriteCursorBlink = true
				}
			}

			for _, z := range zone.GetPrefix(ZoneFileLineNumberPrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFileLineNumberPrefix))
					v.SetCursor(buffer.Point{
						Row: row,
						Col: -1,
					})
					v.SetMark(v.Cursor())
					overwriteCursorBlink = true
				}
			}
		case tea.MouseReleaseMsg:
			for _, z := range append(zone.GetPrefix(ZoneFileDiagnosticPrefix), zone.GetPrefix(ZoneFileLineDiagnosticPrefix)...) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					if !v.Focused() {
						cmds = append(cmds, Focus(ModelTypeFile))
					}

					index, ok := strings.CutPrefix(z.ID(), ZoneFileDiagnosticPrefix)
					if !ok {
						index, _ = strings.CutPrefix(z.ID(), ZoneFileLineDiagnosticPrefix)
					}
					i, _ := strconv.Atoi(index)

					if s := v.Selection(); s != nil && !s.Zero() {
						return v, tea.Batch(cmds...)
					}

					diagnostic := v.file.Diagnostics[i]
					v.SetCursor(diagnostic.Range.Start)
					v.ShowCurrentDiagnostic()
					return v, tea.Batch(cmds...)
				}
			}

			for _, z := range zone.GetPrefix(ZoneFileLinePrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					if !v.Focused() {
						cmds = append(cmds, Focus(ModelTypeFile))
					}

					p := v.GetFileZoneCursorPos(msg, z)
					v.SetCursor(p)
					if s := v.Selection(); s == nil || s.Zero() {
						v.ResetMark()
					}
					cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				case mouse.MatchesZone(msg, z, tea.MouseRight):
					// TODO: open context menu?
					log.Println("right click on line")
					return v, tea.Batch(cmds...)
				}
			}

			for _, z := range zone.GetPrefix(ZoneFileLineNumberPrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					if !v.Focused() {
						cmds = append(cmds, Focus(ModelTypeFile))
					}

					row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFileLineNumberPrefix))
					v.SetCursor(buffer.Point{
						Row: row,
						Col: -1,
					})
					if s := v.Selection(); s == nil || s.Zero() {
						v.ResetMark()
					}
					cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				}
			}

			for _, z := range zone.GetPrefix(ZoneFileLineEmptyPrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					if !v.Focused() {
						cmds = append(cmds, Focus(ModelTypeFile))
					}
					return v, tea.Batch(cmds...)
				}
			}
		case tea.MouseMotionMsg:
			for _, z := range zone.GetPrefix(ZoneFileLinePrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					p := v.GetFileZoneCursorPos(msg, z)
					v.SetCursor(p)
					return v, tea.Batch(cmds...)
				}
			}
		case tea.MouseWheelMsg:
			for _, z := range append(zone.GetPrefix(ZoneFileLinePrefix), zone.GetPrefix(ZoneFileLineNumberPrefix)...) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseWheelLeft), mouse.MatchesZone(msg, z, tea.MouseWheelDown, tea.ModShift):
					v.MoveCursorLeft(1)
					cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				case mouse.MatchesZone(msg, z, tea.MouseWheelRight), mouse.MatchesZone(msg, z, tea.MouseWheelUp, tea.ModShift):
					v.MoveCursorRight(1)
					cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				case mouse.MatchesZone(msg, z, tea.MouseWheelUp):
					v.MoveCursorUp(1)
					cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				case mouse.MatchesZone(msg, z, tea.MouseWheelDown):
					v.MoveCursorDown(1)
					cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				}
			}
		}

	case tea.KeyMsg:
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Show):
				p := v.Cursor()
				cmds = append(cmds, ls.GetAutocompletion(v.Name(), p))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Cancel) && v.file.Autocomplete.Visible():
				v.file.Autocomplete.ClearCompletions()
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Next) && v.file.Autocomplete.Visible():
				v.file.Autocomplete.Next()
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Prev) && v.file.Autocomplete.Visible():
				v.file.Autocomplete.Previous()
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Apply) && v.file.Autocomplete.Visible():
				completion := v.file.Autocomplete.Selected()
				if completion != nil {
					if completion.Text != "" {
						cmds = append(cmds, v.file.Insert(v.Cursor(), []byte(completion.Text)))
					} else if completion.Edit != nil {
						cmds = append(cmds, v.file.Replace(
							completion.Edit.Range,
							[]byte(completion.Edit.NewText)),
						)
					} else {
						cmds = append(cmds, v.file.Insert(v.Cursor(), []byte(completion.Label)))
					}
				}
				v.file.Autocomplete.ClearCompletions()
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.RefreshSyntaxHighlight):
				if cmd := v.file.InitTree(); cmd != nil {
					cmds = append(cmds, cmd)
				}

			case key.Matches(msg, config.Keys.Editor.DebugTreeSitterNodes):
				// TODO: decide where to put this
				// if v.file.Tree == nil {
				//	cmds = append(cmds, notifications.Add("no tree available for this file"))
				//	return v, tea.Batch(cmds...)
				// }
				// buff, err := buffer.New(v.file.Buffer.FileName()+".tree", bytes.NewReader([]byte(v.file.Tree.Print())), "utf-8", buffer.LineEndingLF, false)
				// if err != nil {
				//	cmds = append(cmds, notifications.Add(fmt.Sprintf("error while opening tree.scm: %s", err.Error())))
				//	return v, tea.Batch(cmds...)
				// }
				//
				// debugFile := file.NewFileWithBuffer(buff, file.ModeReadOnly)
				//
				// e.files = append(e.files, debugFile)
				// e.activeFile = len(e.files) - 1
			case key.Matches(msg, config.Keys.Editor.Diagnostic.Show):
				v.ShowCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Cancel) && v.ShowsCurrentDiagnostic():
				v.HideCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Editor.Code.ShowDeclaration):
				cmds = append(cmds, v.file.ShowDeclaration(v.Cursor()))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowDefinitions):
				cmds = append(cmds, v.file.ShowDefinitions(v.Cursor()))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowTypeDefinition):
				cmds = append(cmds, v.file.ShowTypeDefinitions(v.Cursor()))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowImplementation):
				// cmds = append(cmds, f.ShowImplementations())
				// return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowReferences):
				// cmds = append(cmds, f.ShowReferences())
				// return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.OpenOutline):
				cmds = append(cmds, overlay.Open(NewOutlineOverlay(v.file)))

			case key.Matches(msg, config.Keys.Editor.File.Close):
				cmds = append(cmds, CloseAction)

			case key.Matches(msg, config.Keys.Editor.File.Delete):
				cmds = append(cmds, overlay.Open(NewDeleteOverlay([]string{v.Name()})))
			case key.Matches(msg, config.Keys.Editor.File.Rename):
				cmds = append(cmds, overlay.Open(NewRenameOverlay(v.Name())))
			case key.Matches(msg, config.Keys.Editor.Navigation.LineUp):
				v.MoveCursorUp(moveSize)
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.LineDown):
				v.MoveCursorDown(moveSize)
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.CharacterLeft):
				v.MoveCursorLeft(moveSize)
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.CharacterRight):
				v.MoveCursorRight(moveSize)
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.WordUp):
				v.MoveCursorWordUp()
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.WordDown):
				v.MoveCursorWordDown()
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.WordLeft):
				v.SetCursor(v.file.NextWordLeft(v.Cursor()))
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.WordRight):
				v.SetCursor(v.file.NextWordRight(v.Cursor()))
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.PageUp):
				v.MoveCursorUp(pageSize)
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.PageDown):
				v.MoveCursorDown(pageSize)
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.LineStart):
				v.SetCursor(buffer.Point{
					Row: -1,
					Col: 0,
				})
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.LineEnd):
				c := v.Cursor()
				v.SetCursor(buffer.Point{
					Row: -1,
					Col: v.file.Buffer.LineLen(c.Row),
				})
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.FileStart):
				v.SetCursor(buffer.Point{
					Row: 0,
					Col: 0,
				})
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.FileEnd):
				v.SetCursor(buffer.Point{
					Row: v.file.Buffer.LinesLen() - 1,
					Col: v.file.Buffer.LineLen(v.file.Buffer.LinesLen() - 1),
				})
				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.GoTo):
				cmds = append(cmds, overlay.Open(NewGoToOverlay(v.Cursor())))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Edit.Copy):
				selBytes := v.SelectionBytes()
				if len(selBytes) > 0 {
					cmds = append(cmds, Copy(selBytes))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.Paste):
				cmds = append(cmds, Paste)
			case key.Matches(msg, config.Keys.Editor.Edit.Cut):
				s := v.Selection()
				if s != nil {
					cmds = append(cmds, Cut(*s, v.SelectionBytes()))
				}
			case key.Matches(msg, config.Keys.Editor.Selection.SelectLeft):
				v.SelectLeft(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectRight):
				v.SelectRight(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectUp):
				v.SelectUp(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectDown):
				v.SelectDown(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectAll):
				v.SelectAll()

			case key.Matches(msg, config.Keys.Editor.File.Save):
				cmds = append(cmds, SaveFile(v.Name()))
			case key.Matches(msg, config.Keys.Editor.Edit.Tab):
				cmds = append(cmds, v.file.AddTab(v.Cursor().Row))
			case key.Matches(msg, config.Keys.Editor.Edit.RemoveTab):
				cmds = append(cmds, v.file.RemoveTab(v.Cursor().Row))
			case key.Matches(msg, config.Keys.Editor.Edit.Newline):
				v.ResetMark()
				cmds = append(cmds,
					v.file.InsertNewLine(v.Cursor()),
					v.file.Autocomplete.Update(v.Cursor()),
				)
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteRight):
				s := v.Selection()
				if s != nil {
					cmds = append(cmds, v.file.DeleteRange(*s))
					v.ResetMark()
				} else {
					cmds = append(cmds, v.file.DeleteAfter(v.Cursor()))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteLeft):
				s := v.Selection()
				if s != nil {
					cmds = append(cmds, v.file.DeleteRange(*s))
					v.ResetMark()
				} else {
					cmds = append(cmds, v.file.DeleteBefore(v.Cursor()))
				}
				//	c := v.Cursor()
				//	toDelete := v.file.Buffer.Line(c.Row).RuneBytes(c.Col - 1)
				//	cmds = append(cmds, v.file.DeleteBefore(c, 1))
				//	if lang := v.file.Language; lang != nil && len(lang.Config.AutoPairs) > 0 {
				//		row, col = f.Cursor()
				//
				//		fileTree := f.Tree()
				//		if fileTree != nil {
				//			tree := fileTree.FindTree(buffer.Point{
				//				Row: row,
				//				Col: col,
				//			})
				//			if tree != nil {
				//				node := tree.Tree.RootNode().DescendantForRange(sitter.Point{
				//					Row:    uint32(row),
				//					Column: uint32(col),
				//				},
				//					sitter.Point{
				//						Row:    uint32(row),
				//						Column: uint32(col),
				//					},
				//				)
				//				if node != nil && node.Type() == "string" {
				//					log.Println("IN STRING")
				//				}
				//			}
				//		}
				//
				//		for _, pair := range lang.Config.AutoPairs {
				//			if string(toDelete) != pair.Open {
				//				continue
				//			}
				//			closeWidth := ansi.StringWidth(pair.Close)
				//			behindCursor := f.file.Buffer.BytesRange(
				//				buffer.Point{
				//					Row: row,
				//					Col: col,
				//				},
				//				buffer.Point{
				//					Row: row,
				//					Col: col + closeWidth,
				//				},
				//			)
				//			if string(behindCursor) == pair.Close {
				//				cmds = append(cmds, f.Replace(row, col, row, col+closeWidth, nil))
				//				break
				//			}
				//		}
				//	}
			case key.Matches(msg, config.Keys.Editor.Edit.DuplicateLine):
				s := v.Selection()
				if s != nil {
					cmds = append(cmds, v.file.Insert(v.Cursor(), v.SelectionBytes()))
					v.ResetMark()
				} else {
					cmds = append(cmds, v.file.DuplicateLine(v.Cursor().Row))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteWordLeft):
				s := v.Selection()
				if s != nil {
					cmds = append(cmds, v.file.DeleteRange(*s))
					v.ResetMark()
				} else {
					cmds = append(cmds, v.file.DeleteWordLeft(v.Cursor()))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteWordRight):
				s := v.Selection()
				if s != nil {
					cmds = append(cmds, v.file.DeleteRange(*s))
					v.ResetMark()
				} else {
					cmds = append(cmds, v.file.DeleteWordRight(v.Cursor()))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteLine):
				s := v.Selection()
				if s != nil {
					cmds = append(cmds, v.file.DeleteRange(*s))
					v.ResetMark()
				} else {
					cmds = append(cmds, v.file.DeleteLine(v.Cursor().Row))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.ToggleComment):
				// TODO: implement
				// cmds = append(cmds, v.file.ToggleComment())
				overwriteCursorBlink = true

			default:
				k := msg.Key()
				if k.Text == "" || k.Mod.Contains(tea.ModAlt) || k.Mod.Contains(tea.ModCtrl) || k.Mod.Contains(tea.ModMeta) || k.Mod.Contains(tea.ModSuper) || k.Mod.Contains(tea.ModHyper) {
					break
				}

				text := []byte(k.Text)
				if s := v.Selection(); s != nil {
					cmds = append(cmds, v.file.Replace(*s, text))
					v.ResetMark()
				} else {
					cmds = append(cmds, v.file.Insert(v.Cursor(), text))
				}

				cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))

				// handle auto pairs
				if lang := v.file.Language; lang != nil && len(lang.Config.AutoPairs) > 0 {
					for _, pair := range lang.Config.AutoPairs {
						if string(k.Code) == pair.Open {
							c := v.Cursor()
							c.Col += ansi.StringWidth(pair.Open)
							cmds = append(cmds, v.file.Insert(c, []byte(pair.Close)))
							break
						}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	v.cursor.cursor, cmd = v.cursor.cursor.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if v.Cursor() != oldCursor || overwriteCursorBlink {
		cmds = append(cmds, v.CursorBlinkCmd())
	}

	return v, tea.Batch(cmds...)
}

func (v FileView) View(width int, height int, border bool, debug bool) string {
	styles := config.Theme.UI
	borderStyle := func(strs ...string) string { return strings.Join(strs, " ") }
	if border {
		borderStyle = styles.FileView.BorderStyle.Render
	}

	prefixWidth := lipgloss.Width(strconv.Itoa(v.file.Buffer.LinesLen()))
	width = max(width-prefixWidth-styles.FileView.BorderStyle.GetHorizontalFrameSize()-3, 0)

	// debug takes up 4 lines
	if debug {
		height = max(height-4, 0)
	}

	v.refreshCursorViewOffset(width-2, height)
	c := v.Cursor()
	offset := v.cursor.offset
	realCursorRow := c.Row - offset.Row
	realCursorCol := c.Col - offset.Col

	selection := v.Selection()

	var editorCode string
	positions := make([][]buffer.Point, max(height, 0))
	for i := range height {
		ln := i + offset.Row

		var linePositions []buffer.Point

		codeLineStyle := styles.FileView.LineStyle
		codePrefixStyle := styles.FileView.LinePrefixStyle
		codeLineCharStyle := styles.FileView.LineCharStyle
		if ln == c.Row {
			codeLineStyle = styles.FileView.CurrentLineStyle
			codePrefixStyle = styles.FileView.CurrentLinePrefixStyle
			codeLineCharStyle = styles.FileView.CurrentLineCharStyle
		}

		if ln >= v.file.Buffer.LinesLen() {
			editorCode += borderStyle(zone.Mark(zoneFileLineEmptyID(ln), codeLineStyle.Render(codePrefixStyle.Render(strings.Repeat(" ", width))))) + "\n"
			continue
		}

		lineDiagnostic, lineDiagnosticIndex := v.file.HighestLineDiagnostic(ln)

		var prefix string
		if lineDiagnostic.Severity > 0 {
			prefix = zone.Mark(zoneFileLineDiagnosticID(lineDiagnosticIndex), lineDiagnostic.Severity.Icon().Render())
		} else {
			prefix = " "
		}

		prefixLn := strconv.Itoa(ln + 1)
		prefix += zone.Mark(zoneFileLineNumberID(ln), codePrefixStyle.Render(strings.Repeat(" ", prefixWidth-lipgloss.Width(prefixLn))+prefixLn))

		line := v.file.Buffer.Line(ln)
		if line.Len() < offset.Col {
			editorCode += borderStyle(codeLineStyle.Render(prefix)) + "\n"
			continue
		}

		chars := line.RuneStrings()
		var colOffset int
		var codeLine []byte
		// always draw one character off the screen to ensure the cursor is visible
		for ii := range width - prefixWidth + 1 {
			col := ii + offset.Col

			if col <= line.Len() {
				for len(linePositions) <= ii+colOffset {
					linePositions = append(linePositions, buffer.Point{Row: ln, Col: col})
				}
			}

			inSelection := selection != nil && selection.Contains(buffer.Point{Row: ln, Col: col})

			var char string
			if col > len(chars) {
				codeLine = append(codeLine, codeLineCharStyle.Render(" ")...)
				break
			} else if col == len(chars) {
				char = " "
			} else {
				char = chars[col]
			}

			// Replace tabs with spaces
			if char == "\t" {
				char = " "
			}

			style := v.file.HighestMatchStyle(codeLineCharStyle, ln, col)
			style = v.file.HighestLineColDiagnosticStyle(style, ln, col)

			if ln == c.Row && ii == realCursorCol {
				char = v.cursor.cursor.View(char, style)
			} else if inSelection {
				char = styles.FileView.SelectionStyle.Inherit(style).Render(char)
			} else {
				char = style.Render(char)
			}
			codeLine = append(codeLine, char...)

			paddingStyle := codeLineCharStyle
			labelStyle := config.Theme.UI.FileView.InlayHintStyle
			if inSelection {
				paddingStyle = styles.FileView.SelectionStyle.Inherit(paddingStyle)
				labelStyle = styles.FileView.SelectionStyle.Inherit(labelStyle)
			}
			for _, hint := range v.file.InlayHintsForLineCol(ln, col+1) {
				var label string
				if hint.PaddingLeft {
					label += paddingStyle.Render(" ")
				}
				label += labelStyle.Render(hint.Label)
				if hint.PaddingRight {
					label += paddingStyle.Render(" ")
				}
				codeLine = append(codeLine, label...)
				colOffset += lipgloss.Width(label)
			}
		}

		positions[i] = linePositions

		if lineDiagnostic.Severity > 0 && lineDiagnostic.Range.Start.Row == ln {
			lineWidth := ansi.StringWidth(string(codeLine))
			if lineWidth < width {
				diagnosticLine := zone.Mark(zoneFileDiagnosticID(lineDiagnosticIndex), codeLineCharStyle.Render(lineDiagnostic.ShortView(codeLineStyle)))
				codeLine = append(codeLine, diagnosticLine...)
			}
		}

		lineWidth := ansi.StringWidth(string(codeLine))
		if lineWidth < width {
			codeLine = append(codeLine, codeLineCharStyle.Render(strings.Repeat(" ", width-lineWidth))...)
		}

		editorCodeLine := zone.Mark(zoneFileLineID(ln), string(codeLine))

		editorCode += borderStyle(codeLineStyle.Render(prefix+ansi.Truncate(editorCodeLine, width, ""))) + "\n"
	}

	v.file.Positions = positions

	editorCode = strings.TrimSuffix(editorCode, "\n")

	if v.showCurrentDiagnostic {
		diagnostic := v.file.HighestLineColDiagnostic(c.Row, realCursorCol)
		if diagnostic.Severity > 0 {
			editorCode = overlay.PlacePosition(lipgloss.Left, lipgloss.Top, diagnostic.View(width, height), editorCode,
				overlay.WithMarginX(styles.FileView.LinePrefixStyle.GetHorizontalFrameSize()+prefixWidth+1+c.Col),
				overlay.WithMarginY(realCursorRow+1),
			)
		} else {
			v.HideCurrentDiagnostic()
		}
	} else if v.file.Autocomplete.Visible() {
		editorCode = overlay.PlacePosition(lipgloss.Left, lipgloss.Top, v.file.Autocomplete.View(width, height), editorCode,
			overlay.WithMarginX(styles.FileView.LinePrefixStyle.GetHorizontalFrameSize()+prefixWidth+1+c.Col),
			overlay.WithMarginY(realCursorRow+1),
		)
	}

	if debug {
		matches := v.file.MatchesForLineCol(c.Row, realCursorCol)
		slices.Reverse(matches)
		var currentMatches []string
		for _, match := range matches {
			var currentRef string
			if match.ReferenceType != "" {
				currentRef = fmt.Sprintf(" ref: %s", match.ReferenceType)
			}
			currentMatches = append(currentMatches, fmt.Sprintf("%s (%s: [%d, %d] - [%d, %d]%s)", match.Type, match.Source, match.Range.Start.Row, match.Range.Start.Col, match.Range.End.Row, match.Range.End.Col, currentRef))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Matches: %s", strings.Join(currentMatches, ", ")))

		diagnostics := v.file.DiagnosticsForLineCol(c.Row, realCursorCol)
		var currentDiagnostics []string
		for _, diag := range diagnostics {
			currentDiagnostics = append(currentDiagnostics, fmt.Sprintf("%s (%s: %s [%d, %d] - [%d, %d])", diag.Message, diag.Type, diag.Source, diag.Range.Start.Row, diag.Range.Start.Col, diag.Range.End.Row, diag.Range.End.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Diagnostics: %s", strings.Join(currentDiagnostics, ", ")))

		hints := v.file.InlayHintsForLine(c.Row)
		var currentHints []string
		for _, hint := range hints {
			currentHints = append(currentHints, fmt.Sprintf("%s (%s [%d, %d])", hint.Label, hint.Type, hint.Position.Row, hint.Position.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Inlay Hints: %s", strings.Join(currentHints, ", ")))

		var currentDefinitions []string
		for _, definitions := range v.file.Definitions {
			currentDefinitions = append(currentDefinitions, fmt.Sprintf("%s ([%d, %d] - [%d, %d])", definitions.Name, definitions.Range.Start.Row, definitions.Range.Start.Col, definitions.Range.End.Row, definitions.Range.End.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Definitions: %s", strings.Join(currentDefinitions, ", ")))
	}

	return editorCode
}
