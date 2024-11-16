package buffer

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
	pos   int
	len   int
	point Point
}

func (r *Reader) Next() (Char, bool) {
	if r.pos >= r.len {
		return Char{}, false
	}

	p := r.point
	i := r.pos

	runee := r.buf.Rune(r.pos)
	switch runee {
	case '\n':
		r.point.Row++
		r.point.Col = 0
	default:
		r.point.Col++
	}
	r.pos++

	return Char{
		Rune:  runee,
		Point: p,
		Index: i,
	}, true
}

type Char struct {
	Rune  rune
	Point Point
	Index int
}
