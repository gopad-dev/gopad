package file

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbletea"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/internal/bubbles/cursor"
)

var wordBreakers = []rune{' ', '{', '}', '(', ')', '[', ']', '<', '>', '.', ',', ';', ':', '\'', '"', '`', '/', '\\', '|', '-', '=', '+', '*', '&', '^', '%', '$', '#', '@', '!', '~', '?', '\t'}

type Cursor struct {
	row   int
	col   int
	mark  *Mark
	start bool
	end   bool

	cursor cursor.Model

	offsetRow int
	offsetCol int
}

type Mark struct {
	row int
	col int
}

func (f *File) Focus() tea.Cmd {
	return f.cursor.cursor.Focus()
}

func (f *File) Blur() {
	f.cursor.cursor.Blur()
}

func (f *File) Cursor() (int, int) {
	if f.cursor.start {
		return 0, 0
	}

	if f.cursor.end {
		return f.buffer.LinesLen() - 1, f.buffer.LineLen(f.cursor.row)
	}

	return f.cursor.row, min(f.cursor.col, f.buffer.LineLen(f.cursor.row))
}

func (f *File) CursorOffset() (int, int) {
	return f.cursor.offsetRow, f.cursor.offsetCol
}

func (f *File) SetCursor(row, col int) {
	if row > -1 {
		f.cursor.row = min(max(row, 0), f.buffer.LinesLen()-1)
		f.cursor.start = false
		f.cursor.end = false
	}
	if col > -1 {
		cursorRow, _ := f.Cursor()
		f.cursor.col = min(max(col, 0), f.buffer.LineLen(cursorRow))
		f.cursor.start = false
		f.cursor.end = false
	}
}

func (f *File) UpdateCursor(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.cursor.cursor, cmd = f.cursor.cursor.Update(msg)
	return cmd
}

func (f *File) SetCursorBlink(blink bool) {
	f.cursor.cursor.Blink = blink
}

func (f *File) CursorBlinkCmd() tea.Cmd {
	return f.cursor.cursor.BlinkCmd()
}

func (f *File) SetMark(row, col int) {
	f.cursor.mark = &Mark{
		row: row,
		col: col,
	}
}

func (f *File) ResetMark() {
	f.cursor.mark = nil
}

func (f *File) checkMark() {
	if f.cursor.mark != nil {
		cursorRow, cursorCol := f.Cursor()

		if f.cursor.mark.row == cursorRow && f.cursor.mark.col == cursorCol {
			f.cursor.mark = nil
		}
	}
}

func (f *File) Selection() *buffer.Range {
	if f.cursor.mark == nil {
		return nil
	}

	cursorRow, cursorCol := f.Cursor()

	if cursorRow < f.cursor.mark.row || (cursorRow == f.cursor.mark.row && cursorCol < f.cursor.mark.col) {
		return &buffer.Range{
			Start: buffer.Position{
				Row: cursorRow,
				Col: cursorCol,
			},
			End: buffer.Position{
				Row: f.cursor.mark.row,
				Col: f.cursor.mark.col,
			},
		}
	}

	return &buffer.Range{
		Start: buffer.Position{
			Row: f.cursor.mark.row,
			Col: f.cursor.mark.col,
		},
		End: buffer.Position{
			Row: cursorRow,
			Col: cursorCol,
		},
	}
}

func (f *File) SelectionBytes() []byte {
	s := f.Selection()
	if s == nil || s.Start.Row == s.End.Row && s.Start.Col == s.End.Col {
		return nil
	}
	return f.buffer.BytesRange(s.Start, s.End)
}

func (f *File) SelectAll() {
	f.cursor.start = false
	f.cursor.end = false

	f.cursor.mark = &Mark{
		row: 0,
		col: 0,
	}
	f.cursor.row = f.buffer.LinesLen() - 1
	f.cursor.col = f.buffer.LineLen(f.cursor.row)
}

func (f *File) SelectUp(count int) {
	if f.cursor.mark == nil {
		cursorRow, cursorCol := f.Cursor()

		f.cursor.mark = &Mark{
			row: cursorRow,
			col: cursorCol,
		}
	}

	f.moveCursorUp(count)
	f.checkMark()
}

func (f *File) MoveCursorUp(count int) {
	if f.cursor.mark != nil {
		f.cursor.mark = nil
	}

	f.moveCursorUp(count)
}

func (f *File) moveCursorUp(count int) {
	if f.cursor.row == 0 {
		f.cursor.start = true
		return
	}

	f.cursor.end = false
	f.cursor.row = max(0, f.cursor.row-count)
}

func (f *File) SelectDown(count int) {
	if f.cursor.mark == nil {
		cursorRow, cursorCol := f.Cursor()

		f.cursor.mark = &Mark{
			row: cursorRow,
			col: cursorCol,
		}
	}

	f.moveCursorDown(count)
	f.checkMark()
}

func (f *File) MoveCursorDown(count int) {
	if f.cursor.mark != nil {
		f.cursor.mark = nil
	}

	f.moveCursorDown(count)
}

func (f *File) moveCursorDown(count int) {
	if f.cursor.row == f.buffer.LinesLen()-1 {
		f.cursor.end = true
		return
	}

	f.cursor.start = false
	f.cursor.row = min(f.buffer.LinesLen()-1, f.cursor.row+count)
}

func (f *File) SelectLeft(count int) {
	if f.cursor.mark == nil {
		cursorRow, cursorCol := f.Cursor()

		f.cursor.mark = &Mark{
			row: cursorRow,
			col: cursorCol,
		}
	}

	f.moveCursorLeft(count)
	f.checkMark()
}

func (f *File) MoveCursorLeft(count int) {
	if f.cursor.mark != nil {
		f.cursor.mark = nil
	}

	f.moveCursorLeft(count)
}

func (f *File) moveCursorLeft(count int) {
	for range count {
		cursorRow, cursorCol := f.Cursor()

		if cursorCol > 0 {
			f.cursor.col = cursorCol - 1
			f.cursor.start = false
			f.cursor.end = false
		} else if cursorRow > 0 {
			f.cursor.row = cursorRow - 1
			f.cursor.col = f.buffer.LineLen(cursorRow - 1)
			f.cursor.start = false
			f.cursor.end = false
		}
	}
}

func (f *File) SelectRight(count int) {
	if f.cursor.mark == nil {
		cursorRow, cursorCol := f.Cursor()

		f.cursor.mark = &Mark{
			row: cursorRow,
			col: cursorCol,
		}
	}

	f.moveCursorRight(count)
	f.checkMark()
}

func (f *File) MoveCursorRight(count int) {
	if f.cursor.mark != nil {
		f.cursor.mark = nil
	}

	f.moveCursorRight(count)
}

func (f *File) moveCursorRight(count int) {
	for range count {
		cursorRow, cursorCol := f.Cursor()

		if cursorCol < f.buffer.LineLen(cursorRow) {
			f.cursor.col = cursorCol + 1
			f.cursor.start = false
			f.cursor.end = false
		} else if cursorRow < f.buffer.LinesLen()-1 {
			f.cursor.col = 0
			f.cursor.row = cursorRow + 1
			f.cursor.start = false
			f.cursor.end = false
		}
	}
}

func (f *File) MoveCursorWordUp() {
	cursorRow, _ := f.Cursor()

	if cursorRow == 0 {
		return
	}

	var ready bool
	for {
		if cursorRow == 0 || (ready && len(strings.TrimSpace(f.buffer.Line(cursorRow).String())) > 0) {
			break
		}
		cursorRow--
		if !ready {
			ready = true
		}
	}

	f.cursor.row = cursorRow
}

func (f *File) MoveCursorWordDown() {
	cursorRow, _ := f.Cursor()

	if cursorRow == f.buffer.LinesLen()-1 {
		return
	}

	var ready bool
	for {
		if cursorRow == f.buffer.LinesLen()-1 || (ready && len(strings.TrimSpace(f.buffer.Line(cursorRow).String())) > 0) {
			break
		}
		cursorRow++
		if !ready {
			ready = true
		}
	}

	f.cursor.row = cursorRow
}

func (f *File) refreshCursorViewOffset(width int, height int) {
	cursorRow, cursorCol := f.Cursor()

	if cursorRow >= f.cursor.offsetRow+height {
		f.cursor.offsetRow = cursorRow - height + 1
	} else if cursorRow < f.cursor.offsetRow {
		f.cursor.offsetRow = cursorRow
	}

	if cursorCol >= f.cursor.offsetCol+width {
		f.cursor.offsetCol = cursorCol - width + 1
	} else if cursorCol < f.cursor.offsetCol {
		f.cursor.offsetCol = cursorCol
	}
}

func (f *File) NextWordLeft() (int, int) {
	cursorRow, cursorCol := f.Cursor()

	if cursorCol == 0 {
		if cursorRow == 0 {
			return 0, 0
		}

		return cursorRow - 1, f.buffer.LineLen(cursorRow - 1)
	}

	var ready bool
	for {
		if cursorCol == 0 || (ready && slices.Contains(wordBreakers, f.buffer.Line(cursorRow).Rune(cursorCol))) {
			break
		}
		cursorCol--
		if f.buffer.Line(cursorRow).Rune(cursorCol) != ' ' {
			ready = true
		}
	}

	return cursorRow, cursorCol
}

func (f *File) NextWordRight() (int, int) {
	cursorRow, cursorCol := f.Cursor()

	if cursorCol == f.buffer.LineLen(cursorRow) {
		if cursorRow == f.buffer.LinesLen()-1 {
			return cursorRow, cursorCol
		}

		return cursorRow + 1, 0
	}

	var ready bool
	for {
		if cursorCol == f.buffer.LineLen(cursorRow) || (ready && slices.Contains(wordBreakers, f.buffer.Line(cursorRow).Rune(cursorCol))) {
			break
		}
		cursorCol++
		if f.buffer.Line(cursorRow).Rune(cursorCol-1) != ' ' {
			ready = true
		}
	}

	return cursorRow, cursorCol
}
