package editor

import (
	"strings"

	"github.com/charmbracelet/bubbletea/v2"

	"go.gopad.dev/gopad/gopad/editor/buffer"

	"go.gopad.dev/gopad/internal/bubbles/cursor"
)

type fileCursor struct {
	cursor cursor.Model

	point  buffer.Point
	mark   *buffer.Point
	offset buffer.Point

	start bool
	end   bool
}

func (v *DocumentView) Cursor() buffer.Point {
	if v.cursor.start {
		return buffer.Point{
			Row: 0,
			Col: 0,
		}
	}

	if v.cursor.end {
		return buffer.Point{
			Row: v.file.Buffer.LinesLen() - 1,
			Col: v.file.Buffer.LineLen(v.cursor.point.Row),
		}
	}

	return buffer.Point{
		Row: v.cursor.point.Row,
		Col: min(v.cursor.point.Col, v.file.Buffer.LineLen(v.cursor.point.Row)),
	}
}

func (v *DocumentView) SetCursor(newCursor buffer.Point) {
	if newCursor.Row > -1 {
		v.cursor.point.Row = min(max(newCursor.Row, 0), v.file.Buffer.LinesLen()-1)
		v.cursor.start = false
		v.cursor.end = false
	}
	if newCursor.Col > -1 {
		c := v.Cursor()
		v.cursor.point.Col = min(max(newCursor.Col, 0), v.file.Buffer.LineLen(c.Row))
		v.cursor.start = false
		v.cursor.end = false
	}
}

func (v *DocumentView) CursorBlinkCmd() tea.Cmd {
	v.cursor.cursor.Blink = false
	return v.cursor.cursor.BlinkCmd()
}

func (v *DocumentView) SetMark(p buffer.Point) {
	v.cursor.mark = &p
}

func (v *DocumentView) HasMark() bool {
	return v.cursor.mark != nil
}

func (v *DocumentView) ResetMark() {
	v.cursor.mark = nil
}

func (v *DocumentView) checkMark() {
	if v.cursor.mark != nil {
		c := v.Cursor()

		if v.cursor.mark.Row == c.Row && v.cursor.mark.Col == c.Col {
			v.cursor.mark = nil
		}
	}
}

func (v *DocumentView) Selection() *buffer.Range {
	if v.cursor.mark == nil {
		return nil
	}

	c := v.Cursor()
	if c.Row == v.cursor.mark.Row && c.Col == v.cursor.mark.Col {
		return nil
	}

	if c.Row < v.cursor.mark.Row || (c.Row == v.cursor.mark.Row && c.Col < v.cursor.mark.Col) {
		return &buffer.Range{
			Start: buffer.Point{
				Row: c.Row,
				Col: c.Col,
			},
			End: buffer.Point{
				Row: v.cursor.mark.Row,
				Col: v.cursor.mark.Col,
			},
		}
	}

	return &buffer.Range{
		Start: buffer.Point{
			Row: v.cursor.mark.Row,
			Col: v.cursor.mark.Col,
		},
		End: buffer.Point{
			Row: c.Row,
			Col: c.Col,
		},
	}
}

func (v *DocumentView) SelectionBytes() []byte {
	s := v.Selection()
	if s == nil || s.Start.Row == s.End.Row && s.Start.Col == s.End.Col {
		return nil
	}
	return v.file.Buffer.BytesRange(*s)
}

func (v *DocumentView) SelectAll() {
	v.cursor.start = false
	v.cursor.end = false

	v.cursor.mark = &buffer.Point{
		Row: 0,
		Col: 0,
	}
	v.cursor.point.Row = v.file.Buffer.LinesLen() - 1
	v.cursor.point.Col = v.file.Buffer.LineLen(v.cursor.point.Row)
}

func (v *DocumentView) SelectUp(count int) {
	if v.cursor.mark == nil {
		c := v.Cursor()

		v.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	v.moveCursorUp(count)
	v.checkMark()
}

func (v *DocumentView) MoveCursorUp(count int) {
	if v.cursor.mark != nil {
		v.SetCursor(*v.cursor.mark)
		v.ResetMark()
	}

	v.moveCursorUp(count)
}

func (v *DocumentView) moveCursorUp(count int) {
	if v.cursor.point.Row == 0 {
		v.cursor.start = true
		return
	}

	v.cursor.end = false
	v.cursor.point.Row = max(0, v.cursor.point.Row-count)
}

func (v *DocumentView) SelectDown(count int) {
	if v.cursor.mark == nil {
		c := v.Cursor()

		v.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	v.moveCursorDown(count)
	v.checkMark()
}

func (v *DocumentView) MoveCursorDown(count int) {
	if v.cursor.mark != nil {
		v.cursor.mark = nil
	}

	v.moveCursorDown(count)
}

func (v *DocumentView) moveCursorDown(count int) {
	if v.cursor.point.Row == v.file.Buffer.LinesLen()-1 {
		v.cursor.end = true
		return
	}

	v.cursor.start = false
	v.cursor.point.Row = min(v.file.Buffer.LinesLen()-1, v.cursor.point.Row+count)
}

func (v *DocumentView) SelectLeft(count int) {
	if v.cursor.mark == nil {
		c := v.Cursor()

		v.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	v.moveCursorLeft(count)
	v.checkMark()
}

func (v *DocumentView) MoveCursorLeft(count int) {
	if v.cursor.mark != nil {
		v.SetCursor(*v.cursor.mark)
		v.ResetMark()
	}

	v.moveCursorLeft(count)
}

func (v *DocumentView) moveCursorLeft(count int) {
	for range count {
		c := v.Cursor()

		if c.Col > 0 {
			v.cursor.point.Col = c.Col - 1
			v.cursor.start = false
			v.cursor.end = false
		} else if c.Row > 0 {
			v.cursor.point.Row = c.Row - 1
			v.cursor.point.Col = v.file.Buffer.LineLen(c.Row - 1)
			v.cursor.start = false
			v.cursor.end = false
		}
	}
}

func (v *DocumentView) SelectRight(count int) {
	if v.cursor.mark == nil {
		c := v.Cursor()

		v.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	v.moveCursorRight(count)
	v.checkMark()
}

func (v *DocumentView) MoveCursorRight(count int) {
	if v.cursor.mark != nil {
		v.cursor.mark = nil
	}

	v.moveCursorRight(count)
}

func (v *DocumentView) moveCursorRight(count int) {
	for range count {
		c := v.Cursor()

		if c.Col < v.file.Buffer.LineLen(c.Row) {
			v.cursor.point.Col = c.Col + 1
			v.cursor.start = false
			v.cursor.end = false
		} else if c.Row < v.file.Buffer.LinesLen()-1 {
			v.cursor.point.Col = 0
			v.cursor.point.Row = c.Row + 1
			v.cursor.start = false
			v.cursor.end = false
		}
	}
}

func (v *DocumentView) MoveCursorWordUp() {
	c := v.Cursor()

	if c.Row == 0 {
		return
	}

	var ready bool
	for {
		if c.Row == 0 || (ready && len(strings.TrimSpace(v.file.Buffer.Line(c.Row).String())) > 0) {
			break
		}
		c.Row--
		if !ready {
			ready = true
		}
	}

	v.cursor.point.Row = c.Row
}

func (v *DocumentView) MoveCursorWordDown() {
	c := v.Cursor()

	if c.Row == v.file.Buffer.LinesLen()-1 {
		return
	}

	var ready bool
	for {
		if c.Row == v.file.Buffer.LinesLen()-1 || (ready && len(strings.TrimSpace(v.file.Buffer.Line(c.Row).String())) > 0) {
			break
		}
		c.Row++
		if !ready {
			ready = true
		}
	}

	v.cursor.point.Row = c.Row
}
