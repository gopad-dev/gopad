package file

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/lrstanley/bubblezone"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/ls"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/xrunes"
)

type Mode int

const (
	ModeReadOnly Mode = iota
	ModeWrite
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

type Change struct {
	StartIndex  uint32
	OldEndIndex uint32
	NewEndIndex uint32

	Text []byte
}

func NewFileWithBuffer(b *buffer.Buffer, mode Mode) *File {
	f := &File{
		buffer: b,
		mode:   mode,
		cursor: Cursor{
			row:    0,
			col:    0,
			cursor: config.NewCursor(),
		},
		language:           GetLanguageByFilename(b.Name()),
		diagnosticVersions: map[ls.DiagnosticType]int32{},
	}

	f.autocomplete = NewAutocompleter(f)

	return f
}

func NewFileFromName(name string) (*File, error) {
	stat, err := os.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("cannot open directory")
	}

	if stat.Mode().Perm()&0400 == 0 {
		return nil, fmt.Errorf("file is not readable")
	}

	b, err := buffer.NewFromFile(name, "UTF-8", buffer.LineEndingAuto)
	if err != nil {
		return nil, err
	}

	mode := ModeWrite
	if stat.Mode().Perm()&0200 == 0 {
		mode = ModeReadOnly
	}

	return NewFileWithBuffer(b, mode), nil
}

type File struct {
	buffer                *buffer.Buffer
	mode                  Mode
	cursor                Cursor
	language              *Language
	tree                  *Tree
	autocomplete          *Autocompleter
	showCurrentDiagnostic bool

	diagnosticVersions map[ls.DiagnosticType]int32
	diagnostics        []ls.Diagnostic
	inlayHintsVersion  int32
	inlayHints         []ls.InlayHint
	matchesVersion     int32
	matches            [][]*Match
	changes            []Change
	definitions        []ls.Definition
	typeDefinitions    []ls.TypeDefinition
	definitionsIndex   int
	positions          [][]pos
}

func (f *File) Name() string {
	return f.buffer.Name()
}

func (f *File) RelativeName(workspace string) string {
	relName, err := filepath.Rel(workspace, f.Name())
	if err != nil {
		return f.Name()
	}

	return relName
}

func (f *File) FileName() string {
	return f.buffer.FileName()
}

func (f *File) LineEnding() buffer.LineEnding {
	return f.buffer.LineEnding()
}

func (f *File) Encoding() string {
	return f.buffer.EncodingName()
}

func (f *File) Version() int32 {
	return f.buffer.Version()
}

func (f *File) Mode() Mode {
	return f.mode
}

func (f *File) SetMode(mode Mode) {
	f.mode = mode
}

func (f *File) Language() *Language {
	return f.language
}

func (f *File) SetLanguage(name string) {
	language := GetLanguage(name)
	if language == nil {
		return
	}
	f.language = language

	// reset tree and matches when changing language
	f.tree = nil
	f.matches = nil
}

func (f *File) Buffer() *buffer.Buffer {
	return f.buffer
}

func (f *File) Autocomplete() *Autocompleter {
	return f.autocomplete
}

func (f *File) Range() buffer.Range {
	return buffer.Range{
		Start: buffer.Position{Row: 0, Col: 0},
		End:   buffer.Position{Row: f.buffer.LinesLen(), Col: f.buffer.LineLen(max(f.buffer.LinesLen()-1, 0))},
	}
}

func (f *File) Tree() *Tree {
	return f.tree
}

func (f *File) Text() string {
	return f.buffer.String()
}

func (f *File) Dirty() bool {
	return f.buffer.Dirty()
}

func (f *File) recordChange(change Change) tea.Cmd {
	now := time.Now()
	defer func() {
		log.Println("record change time: ", time.Since(now))
	}()

	var cmds []tea.Cmd
	if cmd := f.UpdateTree(sitter.EditInput{
		StartIndex:  change.StartIndex,
		OldEndIndex: change.OldEndIndex,
		NewEndIndex: change.NewEndIndex,
	}); cmd != nil {
		cmds = append(cmds, cmd)
	}

	f.changes = append(f.changes, change)

	cmds = append(cmds, tea.Sequence(
		ls.FileChanged(f.Name(), f.Version(), change.Text),
		ls.GetInlayHint(f.Name(), f.Version(), f.Range()),
	))

	return tea.Batch(cmds...)
}

func (f *File) InsertNewLine() tea.Cmd {
	row, col := f.Cursor()
	startIndex := f.buffer.ByteIndex(row, col)

	f.SetCursor(f.buffer.InsertNewLine(row, col))
	if len(f.matches) > row {
		log.Println("inserting new line")
		for _, lineMatches := range f.matches[row:] {
			for _, match := range lineMatches {
				match.Range.Start.Row++
				match.Range.End.Row++
			}
		}
		f.matches = slices.Insert(f.matches, row, make([]*Match, 0))
	}

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex + 2),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) Insert(text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)
	if len(text) == 0 {
		return nil
	}

	row, col := f.Cursor()
	startIndex := f.buffer.ByteIndex(row, col)

	f.SetCursor(f.buffer.Insert(row, col, text))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex + len(text) + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) InsertRunes(text []rune) tea.Cmd {
	return f.Insert([]byte(string(text)))
}

func (f *File) InsertAt(row int, col int, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)
	if len(text) == 0 {
		return nil
	}

	startIndex := f.buffer.ByteIndex(row, col)

	f.buffer.Insert(row, col, text)

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex + len(text) + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) Replace(fromRow int, fromCol int, toRow int, toCol int, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)

	startIndex := f.buffer.ByteIndex(fromRow, fromCol)
	endIndex := f.buffer.ByteIndex(toRow, toCol)

	f.SetCursor(f.buffer.Replace(fromRow, fromCol, toRow, toCol, text))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(endIndex),
		NewEndIndex: uint32(startIndex + len(text)),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) DuplicateLine() tea.Cmd {
	row, _ := f.Cursor()
	line := f.buffer.Line(row)
	startIndex := f.buffer.ByteIndex(row, 0)

	f.SetCursor(f.buffer.DuplicateLine(row), -1)

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex + line.LenBytes() + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) DeleteLine() tea.Cmd {
	row, _ := f.Cursor()
	line := f.buffer.Line(row)
	startIndex := f.buffer.ByteIndex(row, 0)

	f.SetCursor(f.buffer.DeleteLine(row), -1)

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + line.LenBytes() + 1),
		NewEndIndex: uint32(startIndex + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) DeleteBefore(count int) tea.Cmd {
	row, col := f.Cursor()
	startIndex := f.buffer.ByteIndex(row, col)

	f.SetCursor(f.buffer.DeleteBefore(row, col, count))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex - count),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex - count + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) DeleteAfter(count int) tea.Cmd {
	row, col := f.Cursor()
	startIndex := f.buffer.ByteIndex(row, col)

	f.SetCursor(f.buffer.DeleteAfter(row, col, count))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex + 1 - count),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) DeleteRange(from buffer.Position, to buffer.Position) tea.Cmd {
	startIndex := f.buffer.ByteIndex(from.Row, from.Col)
	endIndex := f.buffer.ByteIndex(to.Row, to.Col)

	f.SetCursor(f.buffer.DeleteRange(from.Row, from.Col, to.Row, to.Col))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(endIndex),
		NewEndIndex: uint32(startIndex),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) DeleteWordLeft() tea.Cmd {
	row, col := f.Cursor()
	startIndex := f.buffer.ByteIndex(row, col)

	wRow, wCol := f.NextWordLeft()
	f.SetCursor(f.buffer.DeleteRange(wRow, wCol, row, col))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) DeleteWordRight() tea.Cmd {
	row, col := f.Cursor()
	startIndex := f.buffer.ByteIndex(row, col)

	wRow, wCol := f.NextWordRight()
	f.SetCursor(f.buffer.DeleteRange(row, col, wRow, wCol))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + 1),
		NewEndIndex: uint32(startIndex + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) RemoveTab() tea.Cmd {
	row, col := f.Cursor()
	line := f.buffer.Line(row)
	startIndex := f.buffer.ByteIndex(row, 0)

	f.SetCursor(f.buffer.RemoveTab(row, col))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + line.LenBytes() + 1),
		NewEndIndex: uint32(startIndex + line.LenBytes() - 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) ToggleComment() tea.Cmd {
	if f.language == nil {
		return nil
	}

	r := f.Selection()
	if r == nil {
		return f.ToggleLineComment()
	}

	return f.ToggleBlockComment(r.Start, r.End)
}

func (f *File) ToggleBlockComment(start buffer.Position, end buffer.Position) tea.Cmd {
	if f.language == nil {
		return nil
	}

	row, col := f.Cursor()

	r := f.Selection()
	startIndex := f.buffer.ByteIndex(r.Start.Row, r.Start.Col)
	endIndex := f.buffer.ByteIndex(r.End.Row, r.End.Col)

	f.SetCursor(f.buffer.ToggleBlockComment(r.Start, r.End, row, col, f.language.Config.BlockCommentTokens))

	var newRuneCount int
	for i := r.Start.Row; i <= r.End.Row; i++ {
		newRuneCount += f.buffer.Line(i).Len()
	}

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(endIndex),
		NewEndIndex: uint32(startIndex + newRuneCount),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) ToggleLineComment() tea.Cmd {
	if f.language == nil || len(f.language.Config.LineCommentTokens) == 0 {
		return nil
	}

	row, col := f.Cursor()
	line := f.buffer.Line(row)
	startIndex := f.buffer.ByteIndex(row, 0)

	f.SetCursor(f.buffer.ToggleLineComment(row, col, f.language.Config.LineCommentTokens))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + line.LenBytes() + 1),
		NewEndIndex: uint32(startIndex + line.LenBytes() - 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f File) GetCursorForCharPos(row int, col int) (int, int) {
	positionRow := max(row-f.cursor.offsetRow, 0)
	if positionRow >= len(f.positions) {
		return max(f.buffer.LinesLen()-1, 0), 0
	}

	linePositions := f.positions[positionRow]
	if col >= len(linePositions) {
		return row, f.buffer.LineLen(row)
	}

	p := linePositions[col]
	return p.row, p.col
}

func (f File) GetFileZoneCursorPos(msg tea.MouseMsg, z *zone.ZoneInfo) (int, int) {
	row, _ := strconv.Atoi(strings.TrimPrefix(z.ID(), ZoneFileLinePrefix))
	col, _ := z.Pos(msg)
	row, col = f.GetCursorForCharPos(row, col)
	return row, col
}

type pos struct {
	row int
	col int
}

func (f *File) View(width int, height int, border bool, debug bool) string {
	styles := config.Theme.UI
	borderStyle := func(strs ...string) string { return strings.Join(strs, " ") }
	if border {
		borderStyle = styles.FileView.BorderStyle.Render
	}

	prefixWidth := lipgloss.Width(strconv.Itoa(f.buffer.LinesLen()))
	width = max(width-prefixWidth+styles.FileView.BorderStyle.GetHorizontalFrameSize()+3, 0)

	// debug takes up 4 lines
	if debug {
		height = max(height-4, 0)
	}

	f.refreshCursorViewOffset(width-2, height)
	cursorRow, cursorCol := f.Cursor()
	offsetRow, offsetCol := f.CursorOffset()
	realCursorRow := cursorRow - offsetRow
	realCursorCol := cursorCol - offsetCol

	selection := f.Selection()

	var editorCode string
	positions := make([][]pos, max(height, 0))
	for i := range height {
		ln := i + offsetRow

		var linePositions []pos

		codeLineStyle := styles.FileView.LineStyle
		codePrefixStyle := styles.FileView.LinePrefixStyle
		codeLineCharStyle := styles.FileView.LineCharStyle
		if ln == cursorRow {
			codeLineStyle = styles.FileView.CurrentLineStyle
			codePrefixStyle = styles.FileView.CurrentLinePrefixStyle
			codeLineCharStyle = styles.FileView.CurrentLineCharStyle
		}

		if ln >= f.buffer.LinesLen() {
			editorCode += borderStyle(zone.Mark(zoneFileLineEmptyID(ln), codeLineStyle.Render(codePrefixStyle.Render(strings.Repeat(" ", width))))) + "\n"
			continue
		}

		lineDiagnostic, lineDiagnosticIndex := f.HighestLineDiagnostic(ln)

		var prefix string
		if lineDiagnostic.Severity > 0 {
			prefix = zone.Mark(zoneFileLineDiagnosticID(lineDiagnosticIndex), lineDiagnostic.Severity.Icon().Render())
		} else {
			prefix = " "
		}

		prefixLn := strconv.Itoa(ln + 1)
		prefix += zone.Mark(zoneFileLineNumberID(ln), codePrefixStyle.Render(strings.Repeat(" ", prefixWidth-lipgloss.Width(prefixLn))+prefixLn))

		line := f.buffer.Line(ln)
		if line.Len() < offsetCol {
			editorCode += borderStyle(codeLineStyle.Render(prefix)) + "\n"
			continue
		}

		chars := line.RuneStrings()
		var colOffset int
		var codeLine []byte
		// always draw one character off the screen to ensure the cursor is visible
		for ii := range width - prefixWidth + 1 {
			col := ii + offsetCol

			if col <= line.Len() {
				for len(linePositions) <= ii+colOffset {
					linePositions = append(linePositions, pos{row: ln, col: col})
				}
			}

			inSelection := selection != nil && selection.Contains(buffer.Position{Row: ln, Col: col})

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

			style := f.HighestMatchStyle(codeLineCharStyle, ln, col)
			style = f.HighestLineColDiagnosticStyle(style, ln, col)

			if ln == cursorRow && ii == realCursorCol {
				char = f.cursor.cursor.View(char, style)
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
			for _, hint := range f.InlayHintsForLineCol(ln, col+1) {
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

	f.positions = positions

	editorCode = strings.TrimSuffix(editorCode, "\n")

	if f.showCurrentDiagnostic {
		diagnostic := f.HighestLineColDiagnostic(cursorRow, realCursorCol)
		if diagnostic.Severity > 0 {
			editorCode = overlay.PlacePosition(lipgloss.Left, lipgloss.Top, diagnostic.View(width, height), editorCode,
				overlay.WithMarginX(styles.FileView.LinePrefixStyle.GetHorizontalFrameSize()+prefixWidth+1+cursorCol),
				overlay.WithMarginY(realCursorRow+1),
			)
		} else {
			f.HideCurrentDiagnostic()
		}
	} else if f.autocomplete.Visible() {
		editorCode = overlay.PlacePosition(lipgloss.Left, lipgloss.Top, f.autocomplete.View(width, height), editorCode,
			overlay.WithMarginX(styles.FileView.LinePrefixStyle.GetHorizontalFrameSize()+prefixWidth+1+cursorCol),
			overlay.WithMarginY(realCursorRow+1),
		)
	}

	if debug {
		matches := f.MatchesForLineCol(cursorRow, realCursorCol)
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

		diagnostics := f.DiagnosticsForLineCol(cursorRow, realCursorCol)
		var currentDiagnostics []string
		for _, diag := range diagnostics {
			currentDiagnostics = append(currentDiagnostics, fmt.Sprintf("%s (%s: %s [%d, %d] - [%d, %d])", diag.Message, diag.Type, diag.Source, diag.Range.Start.Row, diag.Range.Start.Col, diag.Range.End.Row, diag.Range.End.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Diagnostics: %s", strings.Join(currentDiagnostics, ", ")))

		hints := f.InlayHintsForLine(cursorRow)
		var currentHints []string
		for _, hint := range hints {
			currentHints = append(currentHints, fmt.Sprintf("%s (%s [%d, %d])", hint.Label, hint.Type, hint.Position.Row, hint.Position.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Inlay Hints: %s", strings.Join(currentHints, ", ")))

		var currentDefinitions []string
		for _, definitions := range f.definitions {
			currentDefinitions = append(currentDefinitions, fmt.Sprintf("%s ([%d, %d] - [%d, %d])", definitions.Name, definitions.Range.Start.Row, definitions.Range.Start.Col, definitions.Range.End.Row, definitions.Range.End.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Definitions: %s", strings.Join(currentDefinitions, ", ")))
	}

	return editorCode
}
