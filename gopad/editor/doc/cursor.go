package doc

import (
	"strings"

	"go.gopad.dev/gopad/gopad/editor/buffer"
)

func newCursor() Cursor {
	return Cursor{
		point: buffer.Point{},
		mark:  nil,
	}
}

type Cursor struct {
	point buffer.Point
	mark  *buffer.Point

	start bool
	end   bool
}

func (d *Document) Cursor() buffer.Point {
	if d.cursor.start {
		return buffer.Point{
			Row: 0,
			Col: 0,
		}
	}

	if d.cursor.end {
		return buffer.Point{
			Row: d.Buffer.LinesLen() - 1,
			Col: d.Buffer.LineLen(d.cursor.point.Row),
		}
	}

	return buffer.Point{
		Row: d.cursor.point.Row,
		Col: min(d.cursor.point.Col, d.Buffer.LineLen(d.cursor.point.Row)),
	}
}

func (d *Document) SetCursor(newCursor buffer.Point) {
	if newCursor.Row > -1 {
		d.cursor.point.Row = min(max(newCursor.Row, 0), d.Buffer.LinesLen()-1)
		d.cursor.start = false
		d.cursor.end = false
	}
	if newCursor.Col > -1 {
		c := d.Cursor()
		d.cursor.point.Col = min(max(newCursor.Col, 0), d.Buffer.LineLen(c.Row))
		d.cursor.start = false
		d.cursor.end = false
	}
}

func (d *Document) SetMark(p buffer.Point) {
	d.cursor.mark = &p
}

func (d *Document) HasMark() bool {
	return d.cursor.mark != nil
}

func (d *Document) ResetMark() {
	d.cursor.mark = nil
}

func (d *Document) checkMark() {
	if d.cursor.mark != nil {
		c := d.Cursor()

		if d.cursor.mark.Row == c.Row && d.cursor.mark.Col == c.Col {
			d.cursor.mark = nil
		}
	}
}

func (d *Document) Selection() *buffer.Range {
	if d.cursor.mark == nil {
		return nil
	}

	c := d.Cursor()
	if c.Row == d.cursor.mark.Row && c.Col == d.cursor.mark.Col {
		return nil
	}

	if c.Row < d.cursor.mark.Row || (c.Row == d.cursor.mark.Row && c.Col < d.cursor.mark.Col) {
		return &buffer.Range{
			Start: buffer.Point{
				Row: c.Row,
				Col: c.Col,
			},
			End: buffer.Point{
				Row: d.cursor.mark.Row,
				Col: d.cursor.mark.Col,
			},
		}
	}

	return &buffer.Range{
		Start: buffer.Point{
			Row: d.cursor.mark.Row,
			Col: d.cursor.mark.Col,
		},
		End: buffer.Point{
			Row: c.Row,
			Col: c.Col,
		},
	}
}

func (d *Document) SelectionBytes() []byte {
	s := d.Selection()
	if s == nil || s.Start.Row == s.End.Row && s.Start.Col == s.End.Col {
		return nil
	}
	return d.Buffer.BytesRange(*s)
}

func (d *Document) SelectAll() {
	d.cursor.start = false
	d.cursor.end = false

	d.cursor.mark = &buffer.Point{
		Row: 0,
		Col: 0,
	}
	d.cursor.point.Row = d.Buffer.LinesLen() - 1
	d.cursor.point.Col = d.Buffer.LineLen(d.cursor.point.Row)
}

func (d *Document) SelectUp(count int) {
	if d.cursor.mark == nil {
		c := d.Cursor()

		d.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	d.moveCursorUp(count)
	d.checkMark()
}

func (d *Document) MoveCursorUp(count int) {
	if d.cursor.mark != nil {
		d.SetCursor(*d.cursor.mark)
		d.ResetMark()
	}

	d.moveCursorUp(count)
}

func (d *Document) moveCursorUp(count int) {
	if d.cursor.point.Row == 0 {
		d.cursor.start = true
		return
	}

	d.cursor.end = false
	d.cursor.point.Row = max(0, d.cursor.point.Row-count)
}

func (d *Document) SelectDown(count int) {
	if d.cursor.mark == nil {
		c := d.Cursor()

		d.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	d.moveCursorDown(count)
	d.checkMark()
}

func (d *Document) MoveCursorDown(count int) {
	if d.cursor.mark != nil {
		d.cursor.mark = nil
	}

	d.moveCursorDown(count)
}

func (d *Document) moveCursorDown(count int) {
	if d.cursor.point.Row == d.Buffer.LinesLen()-1 {
		d.cursor.end = true
		return
	}

	d.cursor.start = false
	d.cursor.point.Row = min(d.Buffer.LinesLen()-1, d.cursor.point.Row+count)
}

func (d *Document) SelectLeft(count int) {
	if d.cursor.mark == nil {
		c := d.Cursor()

		d.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	d.moveCursorLeft(count)
	d.checkMark()
}

func (d *Document) MoveCursorLeft(count int) {
	if d.cursor.mark != nil {
		d.SetCursor(*d.cursor.mark)
		d.ResetMark()
	}

	d.moveCursorLeft(count)
}

func (d *Document) moveCursorLeft(count int) {
	for range count {
		c := d.Cursor()

		if c.Col > 0 {
			d.cursor.point.Col = c.Col - 1
			d.cursor.start = false
			d.cursor.end = false
		} else if c.Row > 0 {
			d.cursor.point.Row = c.Row - 1
			d.cursor.point.Col = d.Buffer.LineLen(c.Row - 1)
			d.cursor.start = false
			d.cursor.end = false
		}
	}
}

func (d *Document) SelectRight(count int) {
	if d.cursor.mark == nil {
		c := d.Cursor()

		d.cursor.mark = &buffer.Point{
			Row: c.Row,
			Col: c.Col,
		}
	}

	d.moveCursorRight(count)
	d.checkMark()
}

func (d *Document) MoveCursorRight(count int) {
	if d.cursor.mark != nil {
		d.cursor.mark = nil
	}

	d.moveCursorRight(count)
}

func (d *Document) moveCursorRight(count int) {
	for range count {
		c := d.Cursor()

		if c.Col < d.Buffer.LineLen(c.Row) {
			d.cursor.point.Col = c.Col + 1
			d.cursor.start = false
			d.cursor.end = false
		} else if c.Row < d.Buffer.LinesLen()-1 {
			d.cursor.point.Col = 0
			d.cursor.point.Row = c.Row + 1
			d.cursor.start = false
			d.cursor.end = false
		}
	}
}

func (d *Document) MoveCursorWordUp() {
	c := d.Cursor()

	if c.Row == 0 {
		return
	}

	var ready bool
	for {
		if c.Row == 0 || (ready && len(strings.TrimSpace(d.Buffer.Line(c.Row).String())) > 0) {
			break
		}
		c.Row--
		if !ready {
			ready = true
		}
	}

	d.cursor.point.Row = c.Row
}

func (d *Document) MoveCursorWordDown() {
	c := d.Cursor()

	if c.Row == d.Buffer.LinesLen()-1 {
		return
	}

	var ready bool
	for {
		if c.Row == d.Buffer.LinesLen()-1 || (ready && len(strings.TrimSpace(d.Buffer.Line(c.Row).String())) > 0) {
			break
		}
		c.Row++
		if !ready {
			ready = true
		}
	}

	d.cursor.point.Row = c.Row
}
