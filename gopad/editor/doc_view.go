package editor

import (
	"fmt"
	"iter"
	"log"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/bubblezone"

	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/editor/buffer"
	"go.gopad.dev/gopad/gopad/editor/doc"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/key"
	"go.gopad.dev/gopad/internal/bubbles/mouse"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
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

func newDocumentView(name string, buff buffer.Buffer, mode doc.Mode) (*DocumentView, error) {
	f, err := doc.NewDocumentWithBuffer(name, buff, mode)
	if err != nil {
		return nil, err
	}

	return &DocumentView{
		Doc:         f,
		viewOffsetX: -1,
		viewOffsetY: -1,
	}, nil
}

func newDocumentViewFromName(name string) (*DocumentView, error) {
	f, err := doc.NewDocumentFromName(name)
	if err != nil {
		return nil, err
	}
	return &DocumentView{
		Doc:         f,
		viewOffsetX: -1,
		viewOffsetY: -1,
	}, nil
}

type DocumentView struct {
	Doc    *doc.Document
	offset buffer.Point

	lastCursorPosX int
	lastCursorPosY int
	viewOffsetX    int
	viewOffsetY    int

	focused               bool
	showCurrentDiagnostic bool
	definitionsIndex      int
}

func (v *DocumentView) Language() *doc.Language {
	if v.Doc.Syntax == nil {
		return nil
	}
	return v.Doc.Syntax.Language
}

func (v *DocumentView) LanguageName() string {
	language := v.Language()
	if language == nil {
		return ""
	}
	return language.Name
}

func (v *DocumentView) LineEnding() buffer.LineEnding {
	return v.Doc.Buffer.LineEnding()
}

func (v *DocumentView) Focus() tea.Cmd {
	v.focused = true
	return tea.ShowCursor
}

func (v *DocumentView) Blur() tea.Cmd {
	v.focused = false
	//return tea.HideCursor
	return nil
}

func (v DocumentView) Focused() bool {
	return v.focused
}

func (v *DocumentView) ShowsCurrentDiagnostic() bool {
	return v.showCurrentDiagnostic
}

func (v *DocumentView) ShowCurrentDiagnostic() {
	v.showCurrentDiagnostic = true
}

func (v *DocumentView) HideCurrentDiagnostic() {
	v.showCurrentDiagnostic = false
}

func (v DocumentView) GetCursorForCharPos(p buffer.Point) buffer.Point {
	positionRow := max(p.Row-v.offset.Row, 0)
	if positionRow >= len(v.Doc.Positions) {
		return buffer.Point{
			Row: max(v.Doc.Buffer.LinesLen()-1, 0),
			Col: 0,
		}
	}

	linePositions := v.Doc.Positions[positionRow]
	if p.Col >= len(linePositions) {
		return buffer.Point{
			Row: p.Row,
			Col: v.Doc.Buffer.LineLen(p.Row),
		}
	}

	return linePositions[p.Col]
}

func (v DocumentView) GetFileZoneCursorPos(msg tea.MouseMsg, z *zone.ZoneInfo) buffer.Point {
	row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFileLinePrefix))
	col, _ := z.Pos(msg)
	return v.GetCursorForCharPos(buffer.Point{Row: row, Col: col})
}

func (v *DocumentView) SetLanguage(language string) tea.Cmd {
	if err := v.Doc.SetLanguage(language); err != nil {
		return notifications.Add(fmt.Sprintf("failed to set language: %s", err.Error()))
	}
	v.Doc.ClearDiagnosticsByType(ls.DiagnosticTypeTreeSitter)

	return nil
}

func (v *DocumentView) refreshCursorViewOffset(width int, height int) {
	c := v.Doc.Cursor()

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

	if c.Row >= v.offset.Row+height {
		v.offset.Row = c.Row - height + 1
	} else if c.Row < v.offset.Row {
		v.offset.Row = c.Row
	}

	if c.Col >= v.offset.Col+width {
		v.offset.Col = c.Col - width + 1
	} else if c.Col < v.offset.Col {
		v.offset.Col = c.Col
	}
}

func (v DocumentView) Update(msg tea.Msg) (DocumentView, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ls.UpdateFileDiagnosticMsg:
		if msg.Name != v.Doc.Name {
			return v, tea.Batch(cmds...)
		}
		v.Doc.SetDiagnostic(msg.Type, msg.Version, msg.Diagnostics)
		return v, tea.Batch(cmds...)
	case ls.UpdateAutocompletionMsg:
		if msg.Name != v.Doc.Name {
			return v, tea.Batch(cmds...)
		}
		//v.file.Autocomplete.SetCompletions(msg.Completions)
		return v, tea.Batch(cmds...)
	case ls.UpdateInlayHintMsg:
		if msg.Name != v.Doc.Name {
			return v, tea.Batch(cmds...)
		}
		v.Doc.SetInlayHint(msg.Version, msg.Hints)
		return v, tea.Batch(cmds...)
	case ls.RefreshInlayHintMsg:
		cmds = append(cmds, ls.GetInlayHint(v.Doc.Name, v.Doc.Version(), v.Doc.Range()))
		return v, tea.Batch(cmds...)
	case ls.UpdateDeclarationsMsg:
		if msg.Name != v.Doc.Name {
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

		v.Doc.SetDeclarations(msg.Declarations)
		return v, tea.Batch(cmds...)
	case ls.UpdateDefinitionsMsg:
		if msg.Name != v.Doc.Name {
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
		v.Doc.SetDefinitions(msg.Definitions)
		return v, tea.Batch(cmds...)
	case ls.UpdateTypeDefinitionsMsg:
		if msg.Name != v.Doc.Name {
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

		v.Doc.SetTypeDefinitions(msg.TypeDefinitions)
		return v, tea.Batch(cmds...)
	case ls.UpdateImplementationsMsg:
		if msg.Name != v.Doc.Name {
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

		v.Doc.SetImplementations(msg.Implementations)
		return v, tea.Batch(cmds...)
	case ls.UpdateReferencesMsg:
		if msg.Name != v.Doc.Name {
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

		v.Doc.SetReferences(msg.References)
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

					diagnostic := v.Doc.Diagnostics[i]
					v.Doc.SetCursor(diagnostic.Range.Start)
					v.Doc.SetMark(v.Doc.Cursor())
					return v, tea.Batch(cmds...)
				}
			}

			for _, z := range zone.GetPrefix(ZoneFileLinePrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					p := v.GetFileZoneCursorPos(msg, z)
					v.Doc.SetCursor(p)
					v.Doc.SetMark(v.Doc.Cursor())
				}
			}

			for _, z := range zone.GetPrefix(ZoneFileLineNumberPrefix) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseLeft):
					row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFileLineNumberPrefix))
					v.Doc.SetCursor(buffer.Point{
						Row: row,
						Col: -1,
					})
					v.Doc.SetMark(v.Doc.Cursor())
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

					if s := v.Doc.Selection(); s != nil && !s.IsEmpty() {
						return v, tea.Batch(cmds...)
					}

					diagnostic := v.Doc.Diagnostics[i]
					v.Doc.SetCursor(diagnostic.Range.Start)
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
					v.Doc.SetCursor(p)
					if s := v.Doc.Selection(); s == nil || s.IsEmpty() {
						v.Doc.ResetMark()
					}
					//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
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
					v.Doc.SetCursor(buffer.Point{
						Row: row,
						Col: -1,
					})
					if s := v.Doc.Selection(); s == nil || s.IsEmpty() {
						v.Doc.ResetMark()
					}
					//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
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
					v.Doc.SetCursor(p)
					return v, tea.Batch(cmds...)
				}
			}
		case tea.MouseWheelMsg:
			for _, z := range append(zone.GetPrefix(ZoneFileLinePrefix), zone.GetPrefix(ZoneFileLineNumberPrefix)...) {
				switch {
				case mouse.MatchesZone(msg, z, tea.MouseWheelLeft), mouse.MatchesZone(msg, z, tea.MouseWheelDown, tea.ModShift):
					v.Doc.MoveCursorLeft(1)
					//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				case mouse.MatchesZone(msg, z, tea.MouseWheelRight), mouse.MatchesZone(msg, z, tea.MouseWheelUp, tea.ModShift):
					v.Doc.MoveCursorRight(1)
					//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				case mouse.MatchesZone(msg, z, tea.MouseWheelUp):
					v.Doc.MoveCursorUp(1)
					//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				case mouse.MatchesZone(msg, z, tea.MouseWheelDown):
					v.Doc.MoveCursorDown(1)
					//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
					return v, tea.Batch(cmds...)
				}
			}
		}

	case tea.KeyMsg:
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, config.Keys.Editor.Autocomplete.Show):
				p := v.Doc.Cursor()
				cmds = append(cmds, ls.GetAutocompletion(v.Doc.Name, p))
				return v, tea.Batch(cmds...)
			//case key.Matches(msg, config.Keys.Cancel) && v.file.Autocomplete.Visible():
			//	v.file.Autocomplete.ClearCompletions()
			//case key.Matches(msg, config.Keys.Editor.Autocomplete.Next) && v.file.Autocomplete.Visible():
			//	v.file.Autocomplete.Next()
			//case key.Matches(msg, config.Keys.Editor.Autocomplete.Prev) && v.file.Autocomplete.Visible():
			//	v.file.Autocomplete.Previous()
			//case key.Matches(msg, config.Keys.Editor.Autocomplete.Apply) && v.file.Autocomplete.Visible():
			//	completion := v.file.Autocomplete.Selected()
			//	if completion != nil {
			//		if completion.Text != "" {
			//			cmds = append(cmds, v.file.Insert(v.Cursor(), []byte(completion.Text)))
			//		} else if completion.Edit != nil {
			//			cmds = append(cmds, v.file.Replace(
			//				completion.Edit.Range,
			//				[]byte(completion.Edit.NewText)),
			//			)
			//		} else {
			//			cmds = append(cmds, v.file.Insert(v.Cursor(), []byte(completion.Label)))
			//		}
			//	}
			//	v.file.Autocomplete.ClearCompletions()
			//	return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.RefreshSyntaxHighlight):
				if v.Doc.Syntax != nil {
					syntax, err := doc.NewSyntax(v.Doc.Syntax.Language, v.Doc.Buffer.Bytes())
					if err != nil {
						cmds = append(cmds, notifications.Addf("failed to refresh syntax highlight: %s", err.Error()))
						return v, tea.Batch(cmds...)
					}
					v.Doc.Syntax = syntax
					cmds = append(cmds, notifications.Add("Syntax highlight refreshed"))
				}
			case key.Matches(msg, config.Keys.Editor.Diagnostic.Show):
				v.ShowCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Cancel) && v.ShowsCurrentDiagnostic():
				v.HideCurrentDiagnostic()
			case key.Matches(msg, config.Keys.Editor.Code.ShowDeclaration):
				cmds = append(cmds, v.Doc.ShowDeclaration(v.Doc.Cursor()))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowDefinitions):
				cmds = append(cmds, v.Doc.ShowDefinitions(v.Doc.Cursor()))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowTypeDefinition):
				cmds = append(cmds, v.Doc.ShowTypeDefinitions(v.Doc.Cursor()))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowImplementation):
				// cmds = append(cmds, f.ShowImplementations())
				// return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Code.ShowReferences):
				// cmds = append(cmds, f.ShowReferences())
				// return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.OpenOutline):
				cmds = append(cmds, overlay.Open(NewOutlineOverlay(v.Doc)))

			case key.Matches(msg, config.Keys.Editor.File.Close):
				cmds = append(cmds, CloseAction)

			case key.Matches(msg, config.Keys.Editor.File.Delete):
				cmds = append(cmds, overlay.Open(NewDeleteOverlay([]string{v.Doc.Name})))
			case key.Matches(msg, config.Keys.Editor.File.Rename):
				cmds = append(cmds, overlay.Open(NewRenameOverlay(v.Doc.Name)))
			case key.Matches(msg, config.Keys.Editor.Navigation.LineUp):
				v.Doc.MoveCursorUp(moveSize)
			case key.Matches(msg, config.Keys.Editor.Navigation.LineDown):
				v.Doc.MoveCursorDown(moveSize)
			case key.Matches(msg, config.Keys.Editor.Navigation.CharacterLeft):
				v.Doc.MoveCursorLeft(moveSize)
			case key.Matches(msg, config.Keys.Editor.Navigation.CharacterRight):
				v.Doc.MoveCursorRight(moveSize)
			case key.Matches(msg, config.Keys.Editor.Navigation.WordUp):
				v.Doc.MoveCursorWordUp()
			case key.Matches(msg, config.Keys.Editor.Navigation.WordDown):
				v.Doc.MoveCursorWordDown()
			case key.Matches(msg, config.Keys.Editor.Navigation.WordLeft):
				v.Doc.SetCursor(v.Doc.NextWordLeft(v.Doc.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.WordRight):
				v.Doc.SetCursor(v.Doc.NextWordRight(v.Doc.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.PageUp):
				v.Doc.MoveCursorUp(pageSize)
			case key.Matches(msg, config.Keys.Editor.Navigation.PageDown):
				v.Doc.MoveCursorDown(pageSize)
			case key.Matches(msg, config.Keys.Editor.Navigation.LineStart):
				v.Doc.SetCursor(buffer.Point{
					Row: -1,
					Col: 0,
				})
			case key.Matches(msg, config.Keys.Editor.Navigation.LineEnd):
				c := v.Doc.Cursor()
				v.Doc.SetCursor(buffer.Point{
					Row: -1,
					Col: v.Doc.Buffer.LineLen(c.Row),
				})
				//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.FileStart):
				v.Doc.SetCursor(buffer.Point{
					Row: 0,
					Col: 0,
				})
				//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.FileEnd):
				v.Doc.SetCursor(buffer.Point{
					Row: v.Doc.Buffer.LinesLen() - 1,
					Col: v.Doc.Buffer.LineLen(v.Doc.Buffer.LinesLen() - 1),
				})
				//cmds = append(cmds, v.file.Autocomplete.Update(v.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Navigation.GoTo):
				cmds = append(cmds, overlay.Open(NewGoToOverlay(v.Doc.Cursor())))
				return v, tea.Batch(cmds...)
			case key.Matches(msg, config.Keys.Editor.Edit.Copy):
				selBytes := v.Doc.SelectionBytes()
				if len(selBytes) > 0 {
					cmds = append(cmds, Copy(selBytes))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.Paste):
				cmds = append(cmds, Paste)
			case key.Matches(msg, config.Keys.Editor.Edit.Cut):
				s := v.Doc.Selection()
				if s != nil {
					cmds = append(cmds, Cut(*s, v.Doc.SelectionBytes()))
				}
			case key.Matches(msg, config.Keys.Editor.Selection.SelectLeft):
				v.Doc.SelectLeft(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectRight):
				v.Doc.SelectRight(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectUp):
				v.Doc.SelectUp(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectDown):
				v.Doc.SelectDown(moveSize)
			case key.Matches(msg, config.Keys.Editor.Selection.SelectAll):
				v.Doc.SelectAll()

			case key.Matches(msg, config.Keys.Editor.File.Save):
				cmds = append(cmds, SaveFile(v.Doc.Name))
			case key.Matches(msg, config.Keys.Editor.Edit.Tab):
				// cmds = append(cmds, v.file.AddTab(v.Cursor().Row))
			case key.Matches(msg, config.Keys.Editor.Edit.RemoveTab):
				// cmds = append(cmds, v.file.RemoveTab(v.Cursor().Row))
			case key.Matches(msg, config.Keys.Editor.Edit.Newline):
				v.Doc.ResetMark()
				cmds = append(cmds, v.Doc.InsertNewLine(v.Doc.Cursor()))
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteRight):
				s := v.Doc.Selection()
				if s != nil {
					cmds = append(cmds, v.Doc.DeleteRange(*s))
					v.Doc.ResetMark()
				} else {
					cmds = append(cmds, v.Doc.DeleteAfter(v.Doc.Cursor()))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteLeft):
				s := v.Doc.Selection()
				if s != nil {
					cmds = append(cmds, v.Doc.DeleteRange(*s))
					v.Doc.ResetMark()
				} else {
					cmds = append(cmds, v.Doc.DeleteBefore(v.Doc.Cursor()))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DuplicateLine):
				s := v.Doc.Selection()
				if s != nil {
					cmds = append(cmds, v.Doc.Insert(v.Doc.Cursor(), v.Doc.SelectionBytes()))
					v.Doc.ResetMark()
				} else {
					// cmds = append(cmds, v.file.DuplicateLine(v.Cursor().Row))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteWordLeft):
				s := v.Doc.Selection()
				if s != nil {
					// cmds = append(cmds, v.file.DeleteRange(*s))
					v.Doc.ResetMark()
				} else {
					// cmds = append(cmds, v.file.DeleteWordLeft(v.Cursor()))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteWordRight):
				s := v.Doc.Selection()
				if s != nil {
					// cmds = append(cmds, v.file.DeleteRange(*s))
					v.Doc.ResetMark()
				} else {
					// cmds = append(cmds, v.file.DeleteWordRight(v.Cursor()))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.DeleteLine):
				s := v.Doc.Selection()
				if s != nil {
					// cmds = append(cmds, v.file.DeleteRange(*s))
					v.Doc.ResetMark()
				} else {
					// cmds = append(cmds, v.file.DeleteLine(v.Cursor().Row))
				}
			case key.Matches(msg, config.Keys.Editor.Edit.ToggleComment):
				// TODO: implement
				// cmds = append(cmds, v.file.ToggleComment())

			default:
				k := msg.Key()
				if k.Text == "" || k.Mod.Contains(tea.ModAlt) || k.Mod.Contains(tea.ModCtrl) || k.Mod.Contains(tea.ModMeta) || k.Mod.Contains(tea.ModSuper) || k.Mod.Contains(tea.ModHyper) {
					break
				}

				text := []byte(k.Text)
				if s := v.Doc.Selection(); s != nil {
					cmds = append(cmds, v.Doc.Replace(*s, text))
					v.Doc.ResetMark()
				} else {
					cmds = append(cmds, v.Doc.Insert(v.Doc.Cursor(), text))
				}
			}
		}
	}

	c := v.Doc.Cursor()
	if v.lastCursorPosX != c.Row || v.lastCursorPosY != c.Col {
		log.Printf("view offset: %d, %d\n", v.viewOffsetX, v.viewOffsetY)

		cmds = append(cmds,
			tea.SetCursorPosition(c.Col+v.viewOffsetX, c.Row+v.viewOffsetY),
		)

		v.lastCursorPosY = c.Col
		v.lastCursorPosX = c.Row
	}

	return v, tea.Batch(cmds...)
}

func (v *DocumentView) renderLine(ln int, lineCode []byte, prefixWidth int, width int, border bool) string {
	codeLineStyle := config.Theme.UI.FileView.LineStyle
	codePrefixStyle := config.Theme.UI.FileView.LinePrefixStyle
	codeLineCharStyle := config.Theme.UI.FileView.LineCharStyle
	if ln == v.Doc.Cursor().Row {
		codeLineStyle = config.Theme.UI.FileView.CurrentLineStyle
		codePrefixStyle = config.Theme.UI.FileView.CurrentLinePrefixStyle
		codeLineCharStyle = config.Theme.UI.FileView.CurrentLineCharStyle
	}

	borderStyle := func(strs ...string) string { return strings.Join(strs, " ") }
	if border {
		borderStyle = config.Theme.UI.FileView.BorderStyle.Render
	}

	// short circuit if the line is out of bounds
	if ln == -1 {
		return borderStyle("") + "\n"
	}

	lineDiagnostic, lineDiagnosticIndex := v.Doc.HighestLineDiagnostic(ln)

	var prefix string
	if lineDiagnostic.Severity > 0 {
		prefix = zone.Mark(zoneFileLineDiagnosticID(lineDiagnosticIndex), lineDiagnostic.Severity.Icon().Render())
	} else {
		prefix = " "
	}

	prefixLn := strconv.Itoa(ln + 1)
	prefix += zone.Mark(zoneFileLineNumberID(ln), codePrefixStyle.Render(strings.Repeat(" ", prefixWidth-lipgloss.Width(prefixLn))+prefixLn))

	lineWidth := ansi.StringWidth(string(lineCode))
	if lineWidth < width {
		lineCode = append(lineCode, codeLineCharStyle.Render(strings.Repeat(" ", width-lineWidth))...)
	}

	return borderStyle(prefix+codeLineStyle.Render(string(lineCode))) + "\n"
}

func (v *DocumentView) View(width int, height int, border bool, debug bool, offsetX int, offsetY int) string {
	prefixWidth := lipgloss.Width(strconv.Itoa(v.Doc.Buffer.LinesLen()))
	borderWidth := config.Theme.UI.FileView.BorderStyle.GetHorizontalFrameSize()
	width = max(width-prefixWidth-borderWidth-3, 0)

	v.viewOffsetX = offsetX + prefixWidth + borderWidth + 3
	v.viewOffsetY = offsetY

	// debug takes up 4 lines
	if debug {
		height = max(height-4, 0)
	}

	v.refreshCursorViewOffset(width-2, height)
	c := v.Doc.Cursor()
	offset := v.offset
	selection := v.Doc.Selection()

	nextStyle, stop := iter.Pull(v.Doc.HighlightIter(nil))
	defer stop()
	charStyle, ok := nextStyle()

	var (
		editorCode string
		lineCode   []byte
		lastChar   buffer.Char

		// keep track of the style under the cursor for debugging
		cursorCharStyle doc.CharStyle
	)
	r := buffer.NewReader(v.Doc.Buffer, offset)
	for char := range r.All() {
		// stop rendering if we are out of the visible area
		if char.Point.Row < offset.Row || char.Point.Row >= offset.Row+height {
			break
		}

		// only render visible lines
		if char.Point.Col < offset.Col || char.Point.Col >= offset.Col+width {
			continue
		}

		lastChar = char

		codeLineCharStyle := config.Theme.UI.FileView.LineCharStyle
		if char.Point.Row == c.Row {
			codeLineCharStyle = config.Theme.UI.FileView.CurrentLineCharStyle
		}

		if char.Index >= charStyle.End {
			for {
				charStyle, ok = nextStyle()
				if !ok {
					break
				}
				if char.Index < charStyle.End {
					break
				}
			}
		}

		if char.Point.Row == c.Row && char.Point.Col == c.Col {
			cursorCharStyle = charStyle
		}

		style := charStyle.Style.Inherit(codeLineCharStyle)
		style = v.Doc.HighestLineColDiagnosticStyle(style, char.Point.Row, char.Point.Col)

		if char.Rune == '\n' {
			if char.Point.Row == c.Row && char.Point.Col == c.Col {
				cursorCharStyle = charStyle
			} else {
				lineCode = append(lineCode, style.Render(" ")...)
			}

			editorCode += v.renderLine(char.Point.Row, lineCode, prefixWidth, width, border)
			lineCode = nil
			continue
		}

		// replace tabs with spaces for now TODO: handle tabs properly
		if char.Rune == '\t' {
			char.Rune = ' '
		}

		inSelection := selection != nil && selection.Contains(char.Point)

		var renderChar string
		if inSelection {
			renderChar = config.Theme.UI.FileView.SelectionStyle.Inherit(style).Render(string(char.Rune))
		} else {
			renderChar = style.Render(string(char.Rune))
		}

		lineCode = append(lineCode, renderChar...)

		paddingStyle := codeLineCharStyle
		labelStyle := config.Theme.UI.FileView.InlayHintStyle
		if inSelection {
			paddingStyle = config.Theme.UI.FileView.SelectionStyle.Inherit(paddingStyle)
			labelStyle = config.Theme.UI.FileView.SelectionStyle.Inherit(labelStyle)
		}
		for _, hint := range v.Doc.InlayHintsForLineCol(char.Point.Row, char.Point.Col+1) {
			var label string
			if hint.PaddingLeft {
				label += paddingStyle.Render(" ")
			}
			label += labelStyle.Render(hint.Label)
			if hint.PaddingRight {
				label += paddingStyle.Render(" ")
			}
			lineCode = append(lineCode, label...)
		}
	}

	if len(lineCode) > 0 {
		editorCode += v.renderLine(lastChar.Point.Row, lineCode, prefixWidth, width, border)
	}

	codeHeight := lipgloss.Height(editorCode)
	if codeHeight <= height {
		editorCode += strings.Repeat(v.renderLine(-1, nil, 0, width, border), height-codeHeight+1)
	}

	editorCode = strings.TrimSuffix(editorCode, "\n")

	if v.showCurrentDiagnostic {
		diagnostic := v.Doc.HighestLineColDiagnostic(c.Row, c.Col)
		if diagnostic.Severity > 0 {
			editorCode = overlay.PlacePosition(lipgloss.Left, lipgloss.Top, diagnostic.View(width, height), editorCode,
				overlay.WithMarginX(config.Theme.UI.FileView.LinePrefixStyle.GetHorizontalFrameSize()+prefixWidth+1+c.Col),
				overlay.WithMarginY(c.Row+1),
			)
		} else {
			v.HideCurrentDiagnostic()
		}
	}

	if debug {
		editorCode += "\n" + fmt.Sprintf("  Cursor Char Style: %s (%s) [%d, %d]", cursorCharStyle.StyleName, cursorCharStyle.LanguageName, cursorCharStyle.Start, cursorCharStyle.End)

		diagnostics := v.Doc.DiagnosticsForLineCol(c.Row, c.Col)
		var currentDiagnostics []string
		for _, diag := range diagnostics {
			currentDiagnostics = append(currentDiagnostics, fmt.Sprintf("%s (%s: %s [%d, %d] - [%d, %d])", diag.Message, diag.Type, diag.Source, diag.Range.Start.Row, diag.Range.Start.Col, diag.Range.End.Row, diag.Range.End.Col))
		}
		editorCode += "\n" + fmt.Sprintf("  Current Diagnostics: %s", strings.Join(currentDiagnostics, ", "))

		hints := v.Doc.InlayHintsForLine(c.Row)
		var currentHints []string
		for _, hint := range hints {
			currentHints = append(currentHints, fmt.Sprintf("%s (%s [%d, %d])", hint.Label, hint.Type, hint.Position.Row, hint.Position.Col))
		}
		editorCode += "\n" + fmt.Sprintf("  Current Inlay Hints: %s", strings.Join(currentHints, ", "))

		var currentDefinitions []string
		for _, definitions := range v.Doc.Definitions {
			currentDefinitions = append(currentDefinitions, fmt.Sprintf("%s ([%d, %d] - [%d, %d])", definitions.Name, definitions.Range.Start.Row, definitions.Range.Start.Col, definitions.Range.End.Row, definitions.Range.End.Col))
		}
		editorCode += "\n" + fmt.Sprintf("  Current Definitions: %s", strings.Join(currentDefinitions, ", "))
	}

	return editorCode
}
