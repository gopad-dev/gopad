package buffer

import (
	"iter"
)

var newLine = []byte("\n")

func NewReader(buf Buffer, offset Point) *Reader {
	return &Reader{
		buf:   buf,
		pos:   buf.ByteIndexByPoint(offset),
		len:   buf.Len(),
		point: offset,
	}
}

type Reader struct {
	buf   Buffer
	pos   uint
	len   uint
	point Point
}

func (r *Reader) All() iter.Seq[Char] {
	return func(yield func(Char) bool) {
		for {
			c, ok := r.next()
			if !ok {
				break
			}
			if !yield(c) {
				break
			}
		}
	}
}

func (r *Reader) next() (Char, bool) {
	if r.pos >= r.len {
		return Char{}, false
	}

	p := r.point
	i := r.pos

	text := r.buf.Rune(r.pos)
	switch text {
	case newLine:
		r.point.Row++
		r.point.Col = 0
	default:
		r.point.Col++
	}
	r.pos++

	return Char{
		Text:  text,
		Point: p,
		Index: i,
	}, true
}

type Char struct {
	Text  []byte
	Point Point
	Index uint
}
