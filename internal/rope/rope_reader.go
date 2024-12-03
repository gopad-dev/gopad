package rope

import (
	"io"
)

func NewReader(rope Rope) *Reader {
	return rope.Reader()
}

type Reader struct {
	rope     Rope
	position int64
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.rope.ReadAt(p, r.position)
	if err == nil {
		r.position += int64(n)
	}
	return
}

func (r Rope) Reader() *Reader {
	return r.OffsetReader(0)
}

func (r Rope) OffsetReader(offset int) *Reader {
	return &Reader{rope: r, position: int64(offset)}
}

func (r Rope) ReadAt(p []byte, off int64) (n int, err error) {
	o := int(off)
	for n < len(p) && o+n < r.Len() {
		leaf, at := r.leafForOffset(o + n)
		n += copy(p[n:], leaf.content[at:])
	}

	if n < len(p) {
		err = io.EOF
	}

	return
}
