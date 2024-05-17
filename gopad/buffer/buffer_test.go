package buffer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer_BytesRange(t *testing.T) {
	r := bytes.NewReader([]byte("hello\nworld\n"))
	b, err := New("test.txt", r, "utf-8", LineEndingLF, false)
	assert.NoError(t, err)

	data := []struct {
		from Position
		to   Position
		want []byte
	}{
		{
			from: Position{
				Row: 0,
				Col: 1,
			},
			to: Position{
				Row: 0,
				Col: 2,
			},
			want: []byte("e"),
		},
		{
			from: Position{
				Row: 0,
				Col: 0,
			},
			to: Position{
				Row: 0,
				Col: 1,
			},
			want: []byte("h"),
		},
	}

	for _, d := range data {
		got := b.BytesRange(d.from, d.to)
		assert.Equal(t, d.want, got)
	}
}

func TestBuffer_Replace(t *testing.T) {
	r := bytes.NewReader([]byte("lol\n()\n"))
	b, err := New("test.txt", r, "utf-8", LineEndingLF, false)
	assert.NoError(t, err)

	data := []struct {
		from Position
		to   Position
		text []byte
		want []byte
	}{
		{
			from: Position{
				Row: 1,
				Col: 0,
			},
			to: Position{
				Row: 1,
				Col: 2,
			},
			text: nil,
			want: []byte("lol\n\n"),
		},
	}

	for _, d := range data {
		b.Replace(d.from.Row, d.from.Col, d.to.Row, d.to.Col, d.text)
		assert.Equal(t, d.want, b.Bytes())
	}
}
