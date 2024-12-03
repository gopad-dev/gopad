package buffer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var source = []byte("hello\nworld\n")

func TestLineBuffer_RuneIndex(t *testing.T) {
	b, err := New(bytes.NewReader([]byte("hello\nworld\n")), LineEndingLF)
	require.NoError(t, err)

	tests := []struct {
		name     string
		index    int
		expected int
	}{
		{
			name:     "start",
			index:    0,
			expected: 0,
		},
		{
			name:     "middle",
			index:    6,
			expected: 6,
		},
		{
			name:     "end",
			index:    11,
			expected: 11,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := b.RuneIndex(test.index)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestLineBuffer_Rune(t *testing.T) {
	b, err := New(bytes.NewReader([]byte("hello\nworld\n")), LineEndingLF)
	require.NoError(t, err)

	tests := []struct {
		name     string
		index    int
		expected rune
	}{
		{
			name:     "start",
			index:    0,
			expected: 'h',
		},
		{
			name:     "middle",
			index:    6,
			expected: 'w',
		},
		{
			name:     "end",
			index:    11,
			expected: '\n',
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := b.Rune(test.index)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestLineBuffer_BytesRange(t *testing.T) {
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

func TestLineBuffer_Replace(t *testing.T) {
	r := bytes.NewReader([]byte("lol\n()\n"))
	b, err := New(r, LineEndingLF)
	assert.NoError(t, err)

	data := []struct {
		from     Point
		to       Point
		text     []byte
		expected []byte
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
			text:     nil,
			expected: []byte("lol\n\n"),
		},
	}

	for _, d := range data {
		b.Replace(Range{
			Start: d.from,
			End:   d.to,
		}, d.text)
		assert.Equal(t, d.expected, b.Bytes())
	}
}

func TestLineBuffer_Insert(t *testing.T) {
	b, err := New(bytes.NewReader(source), LineEndingLF)
	require.NoError(t, err)

	data := []struct {
		at       Point
		text     []byte
		expected []byte
	}{
		{
			at: Point{
				Row: 1,
				Col: 0,
			},
			text:     []byte("hi"),
			expected: []byte("hello\nhiworld\n"),
		},
	}

	for _, d := range data {
		b.Insert(d.at, d.text)
		assert.Equal(t, d.expected, b.Bytes())
	}
}

func TestLineBuffer_Delete(t *testing.T) {
	b, err := New(bytes.NewReader(source), LineEndingLF)
	require.NoError(t, err)

	data := []struct {
		from     Point
		to       Point
		expected []byte
	}{
		{
			from:     Point{Row: 1, Col: 0},
			to:       Point{Row: 1, Col: 5},
			expected: []byte("hello\n\n"),
		},
	}

	for _, d := range data {
		b.Delete(Range{
			Start: d.from,
			End:   d.to,
		})
		assert.Equal(t, d.expected, b.Bytes())
	}
}
