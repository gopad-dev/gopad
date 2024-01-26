package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"go.gopad.dev/go-tree-sitter"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
	"go.gopad.dev/gopad/gopad/lsp"
	"go.gopad.dev/gopad/internal/bubbles/notifications"
	"go.gopad.dev/gopad/internal/bubbles/overlay"
	"go.gopad.dev/gopad/internal/xrunes"
)

type FileMode int

const (
	FileModeReadOnly FileMode = iota
	FileModeWrite
)

type Change struct {
	StartIndex  uint32
	OldEndIndex uint32
	NewEndIndex uint32

	Text []byte
}

func NewFileWithBuffer(b *buffer.Buffer, mode FileMode) *File {
	return &File{
		buffer: b,
		mode:   mode,
		cursor: Cursor{
			row:    0,
			col:    0,
			cursor: config.NewCursor(),
		},
		language:     GetLanguageByFilename(b.Name()),
		autocomplete: NewAutocompleter(),
	}
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

	mode := FileModeWrite
	if stat.Mode().Perm()&0200 == 0 {
		mode = FileModeReadOnly
	}

	return NewFileWithBuffer(b, mode), nil
}

type File struct {
	buffer            *buffer.Buffer
	mode              FileMode
	cursor            Cursor
	language          *Language
	tree              *Tree
	autocomplete      *Autocompleter
	diagnosticVersion int32
	diagnostics       []lsp.Diagnostic
	matches           []Match
	locals            []Local
	changes           []Change
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

func (f *File) LineEnding() any {
	return f.buffer.LineEnding()
}

func (f *File) Encoding() string {
	return f.buffer.EncodingName()
}

func (f *File) Version() int32 {
	return f.buffer.Version()
}

func (f *File) Mode() FileMode {
	return f.mode
}

func (f *File) SetMode(mode FileMode) {
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

func (f *File) SetDiagnostic(version int32, diagnostics []lsp.Diagnostic) {
	if version < f.diagnosticVersion {
		return
	}
	if version > f.diagnosticVersion {
		f.diagnostics = diagnostics
		f.diagnosticVersion = version
		return
	}
	f.diagnostics = append(f.diagnostics, diagnostics...)
}

func (f *File) Diagnostics() []lsp.Diagnostic {
	return f.diagnostics
}

func (f *File) ClearDiagnosticsByType(dType lsp.DiagnosticType) {
	diagnostics := f.diagnostics[:0]
	for i := range f.diagnostics {
		if f.diagnostics[i].Type != dType {
			diagnostics = append(diagnostics, f.diagnostics[i])
		}
	}

	f.diagnostics = diagnostics
}

func (f *File) DiagnosticsForLineCol(row int, col int) []lsp.Diagnostic {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostics []lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) {
			diagnostics = append(diagnostics, diag)
		}
	}
	return diagnostics
}

func (f *File) HighestLineDiagnostic(row int) lsp.Diagnostic {
	var diagnostic lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Start.Row == row && (diag.Severity > diagnostic.Severity || (diag.Severity >= diagnostic.Severity && diag.Priority > diagnostic.Priority)) {
			diagnostic = diag
		}
	}
	return diagnostic
}

func (f *File) HighestLineColDiagnostic(row int, col int) lsp.Diagnostic {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostic lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) && (diag.Severity > diagnostic.Severity || (diag.Severity >= diagnostic.Severity && diag.Priority > diagnostic.Priority)) {
			diagnostic = diag
		}
	}

	return diagnostic
}

func (f *File) HighestLineColDiagnosticStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostic lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) && (diag.Severity > diagnostic.Severity || (diag.Severity >= diagnostic.Severity && diag.Priority > diagnostic.Priority)) {
			diagnostic = diag
		}
	}

	if diagnostic.Severity == 0 {
		return style
	}

	return diagnostic.Severity.CharStyle().Copy().Inherit(style)
}

func (f *File) Matches() []Match {
	return f.matches
}

func (f *File) MatchesForLineCol(row int, col int) []Match {
	pos := buffer.Position{Row: row, Col: col}

	var matches []Match
	for _, match := range f.matches {
		if match.Range.Contains(pos) {
			matches = append(matches, match)
		}
	}
	return matches
}

func (f *File) HighestMatchStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	for _, match := range f.MatchesForLineCol(row, col) {
		matchType := match.Type
		for {
			codeStyle, ok := config.Theme.Editor.CodeStyles[matchType]
			if ok {
				return style.Copy().Inherit(codeStyle)
			}
			lastDot := strings.LastIndex(matchType, ".")
			if lastDot == -1 {
				break
			}
			matchType = matchType[:lastDot]
		}
	}

	return style
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
	var cmds []tea.Cmd
	if err := f.UpdateTree(sitter.EditInput{
		StartIndex:  change.StartIndex,
		OldEndIndex: change.OldEndIndex,
		NewEndIndex: change.NewEndIndex,
	}); err != nil {
		cmds = append(cmds, notifications.Add(fmt.Sprintf("Error updating tree: %v", err)))
	}

	f.changes = append(f.changes, change)

	cmds = append(cmds, lsp.FileChanged(f.Name(), f.Version(), change.Text))

	return tea.Batch(cmds...)
}

func (f *File) InsertNewLine() tea.Cmd {
	row, col := f.Cursor()
	startIndex := f.buffer.ByteIndex(row, col)

	f.SetCursor(f.buffer.InsertNewLine(row, col))

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
		NewEndIndex: uint32(startIndex + runewidth.StringWidth(string(text)) + 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) InsertRunes(text []rune) tea.Cmd {
	return f.Insert([]byte(string(text)))
}

func (f *File) Replace(fromRow int, fromCol int, toRow int, toCol int, text []byte) tea.Cmd {
	text = xrunes.Sanitize(text)
	if len(text) == 0 {
		return nil
	}

	startIndex := f.buffer.ByteIndex(fromRow, fromCol)
	endIndex := f.buffer.ByteIndex(toRow, toCol)

	f.SetCursor(f.buffer.Replace(fromRow, fromCol, toRow, toCol, text))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(endIndex),
		NewEndIndex: uint32(startIndex + runewidth.StringWidth(string(text))),
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

	f.SetCursor(f.buffer.ToggleBlockComment(r.Start, r.End, row, col, f.language.BlockCommentTokens))

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
	if f.language == nil || len(f.language.LineCommentTokens) == 0 {
		return nil
	}

	row, col := f.Cursor()
	line := f.buffer.Line(row)
	startIndex := f.buffer.ByteIndex(row, 0)

	f.SetCursor(f.buffer.ToggleLineComment(row, col, f.language.LineCommentTokens))

	return f.recordChange(Change{
		StartIndex:  uint32(startIndex),
		OldEndIndex: uint32(startIndex + line.LenBytes() + 1),
		NewEndIndex: uint32(startIndex + line.LenBytes() - 1),
		Text:        f.buffer.Bytes(),
	})
}

func (f *File) View(width int, height int, border bool, debug bool) string {
	styles := config.Theme.Editor
	borderStyle := func(strs ...string) string { return strings.Join(strs, " ") }
	if border {
		borderStyle = styles.CodeBorderStyle.Render
	}

	prefixLength := lipgloss.Width(strconv.Itoa(f.buffer.LinesLen()))
	width -= prefixLength + styles.CodePrefixStyle.GetHorizontalFrameSize() + 1

	if debug {
		height = max(height-2, 0)
	}

	f.refreshCursorViewOffset(width-2, height)
	cursorRow, cursorCol := f.Cursor()
	offsetRow, offsetCol := f.CursorOffset()
	realCursorRow := cursorRow - offsetRow
	realCursorCol := cursorCol - offsetCol

	selection := f.Selection()

	var editorCode string
	for i := range height {
		ln := i + offsetRow

		codeLineStyle := styles.CodeLineStyle
		codePrefixStyle := styles.CodePrefixStyle
		codeLineCharStyle := styles.CodeLineCharStyle
		if ln == cursorRow {
			codeLineStyle = styles.CodeCurrentLineStyle
			codePrefixStyle = styles.CodeCurrentLinePrefixStyle
			codeLineCharStyle = styles.CodeCurrentLineCharStyle
		}

		if ln >= f.buffer.LinesLen() {
			editorCode += borderStyle(codeLineStyle.Render(codePrefixStyle.Render(strings.Repeat(" ", prefixLength)))) + "\n"
			continue
		}

		lineDiagnostic := f.HighestLineDiagnostic(ln)

		var prefix string
		if lineDiagnostic.Severity > 0 {
			prefix = lineDiagnostic.Severity.Style().Render(lineDiagnostic.Severity.ShortString())
		} else {
			prefix = " "
		}

		prefixLn := strconv.Itoa(ln + 1)
		prefix += codePrefixStyle.Render(strings.Repeat(" ", prefixLength-lipgloss.Width(prefixLn)) + prefixLn)

		line := f.buffer.Line(ln)
		if line.Len() < offsetCol {
			editorCode += borderStyle(codeLineStyle.Render(prefix)) + "\n"
			continue
		}

		chars := line.RuneStrings()
		var codeLine []byte
		// always draw one character off the screen to ensure the cursor is visible
		for ii := range width - prefixLength + 1 {
			col := ii + offsetCol
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
			if col != len(chars) {
				style = f.HighestLineColDiagnosticStyle(style, ln, col)
			}

			if selection != nil && selection.Contains(buffer.Position{Row: ln, Col: col}) {
				char = styles.CodeSelectionStyle.Copy().Inherit(style).Render(char)
			} else if ln == cursorRow && ii == realCursorCol {
				char = f.cursor.cursor.View(char, style)
			} else {
				char = style.Render(char)
			}
			codeLine = append(codeLine, char...)
		}

		if lineDiagnostic.Severity > 0 {
			lineWidth := ansi.PrintableRuneWidth(string(codeLine))
			if lineWidth < width {
				codeLine = append(codeLine, codeLineCharStyle.Render(lineDiagnostic.Severity.Style().Render(lineDiagnostic.Message))...)
			}
		}

		lineWidth := ansi.PrintableRuneWidth(string(codeLine))
		if lineWidth < width {
			codeLine = append(codeLine, codeLineCharStyle.Render(strings.Repeat(" ", width-lineWidth))...)
		}

		editorCode += borderStyle(codeLineStyle.Render(prefix+string(codeLine))) + "\n"
	}

	editorCode = strings.TrimSuffix(editorCode, "\n")

	if f.autocomplete.Visible() {
		editorCode = overlay.PlacePosition(lipgloss.Left, lipgloss.Top, f.autocomplete.View(width, height), editorCode,
			overlay.WithMarginX(styles.CodePrefixStyle.GetHorizontalFrameSize()+prefixLength+1+cursorCol),
			overlay.WithMarginY(realCursorRow+1),
		)
	}

	if debug {
		matches := f.MatchesForLineCol(cursorRow, realCursorCol)
		var currentMatches []string
		for _, match := range matches {
			currentMatches = append(currentMatches, fmt.Sprintf("%s (%s: %d-%d) ", match.Type, match.Source, match.Range.Start.Col, match.Range.End.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Matches: %s", strings.Join(currentMatches, ", ")))

		diagnostics := f.DiagnosticsForLineCol(cursorRow, realCursorCol)
		var currentDiagnostics []string
		for _, diag := range diagnostics {
			currentDiagnostics = append(currentDiagnostics, fmt.Sprintf("%s (%s: %s [%d, %d] - [%d, %d]) ", diag.Message, diag.Type, diag.Source, diag.Range.Start.Row, diag.Range.Start.Col, diag.Range.End.Row, diag.Range.End.Col))
		}
		editorCode += "\n" + borderStyle(fmt.Sprintf("  Current Diagnostics: %s", strings.Join(currentDiagnostics, ", ")))
	}

	return editorCode
}
