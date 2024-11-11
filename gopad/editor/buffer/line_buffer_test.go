package buffer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer_BytesRange(t *testing.T) {
	r := bytes.NewReader([]byte("hello\nworld\n"))
	b, err := New(r, LineEndingLF)
	assert.NoError(t, err)

	data := []struct {
		from Point
		to   Point
		want []byte
	}{
		{
			from: Point{
				Row: 0,
				Col: 1,
			},
			to: Point{
				Row: 0,
				Col: 2,
			},
			want: []byte("e"),
		},
		{
			from: Point{
				Row: 0,
				Col: 0,
			},
			to: Point{
				Row: 0,
				Col: 1,
			},
			want: []byte("h"),
		},
	}

	for _, d := range data {
		got := b.BytesRange(Range{
			Start: d.from,
			End:   d.to,
		})
		assert.Equal(t, d.want, got)
	}
}

func TestBuffer_Replace(t *testing.T) {
	r := bytes.NewReader([]byte("lol\n()\n"))
	b, err := New(r, LineEndingLF)
	assert.NoError(t, err)

	data := []struct {
		from Point
		to   Point
		text []byte
		want []byte
	}{
		{
			from: Point{
				Row: 1,
				Col: 0,
			},
			to: Point{
				Row: 1,
				Col: 2,
			},
			text: nil,
			want: []byte("lol\n\n"),
		},
	}

	for _, d := range data {
		b.Replace(Range{
			Start: d.from,
			End:   d.to,
		}, d.text)
		assert.Equal(t, d.want, b.Bytes())
	}
}
