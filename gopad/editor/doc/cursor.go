package doc

import (
	"strings"

	"go.gopad.dev/gopad/gopad/editor/buffer"
)

type Selection struct {
	// Anchor is the starting point of the selection which doesn't move.
	Anchor uint
	// Head is the ending point of the selection which moves when the cursor moves.
	Head uint
}

func newCursor() Cursor {
	return Cursor{
		point: 0,
		mark:  nil,
	}
}

type Cursor struct {
	point uint
	mark  *uint

	start bool
	end   bool
}

func (c Cursor) mapChanges(changes ChangeSet) Cursor {
	if changes.IsEmpty() {
		return c
	}

	selection := Selection{
		Anchor: c.point,
		Head:   c.point,
	}
	if c.mark != nil {
		selection.Anchor = *c.mark
	}

	selections := changes.UpdatePosition([]Selection{selection})

	newC := Cursor{
		point: selections[0].Head,
		mark:  nil,
	}

	if selections[0].Anchor != selections[0].Head {
		newC.mark = &selections[0].Anchor
	}

	return newC
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

func (d *Document) SetCursor(cursor uint) {
	d.cursor.point = cursor
	d.cursor.start = false
	d.cursor.end = false
}

func (d *Document) SetMark(mark uint) {
	d.cursor.mark = &mark
}

func (d *Document) HasMark() bool {
	return d.cursor.mark != nil
}

func (d *Document) ResetMark() {
	d.cursor.mark = nil
}

func (d *Document) checkMark() {
	if d.cursor.mark != nil {
		if *d.cursor.mark == d.cursor.point {
			d.cursor.mark = nil
		}
	}
}

func (d *Document) Selection() *buffer.ByteRange {
	if d.cursor.mark == nil {
		return nil
	}

	return &buffer.ByteRange{
		StartByte: min(*d.cursor.mark, d.cursor.point),
		EndByte:   max(*d.cursor.mark, d.cursor.point),
	}
}

func (d *Document) SelectionBytes() []byte {
	s := d.Selection()
	if s == nil || s.Start.Row == s.End.Row && s.Start.Col == s.End.Col {
		return nil
	}
	return d.Buffer.BytesRange()
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
