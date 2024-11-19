package xbytes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteIndex(t *testing.T) {
	testString := "Hello, 世界"

	tests := []struct {
		name     string
		data     []byte
		index    int
		expected int
	}{
		{
			name:     "start",
			data:     []byte(testString),
			index:    0,
			expected: 0,
		},
		{
			name:     "second rune",
			data:     []byte(testString),
			index:    1,
			expected: 1,
		},
		{
			name:     "out of bounds",
			data:     []byte(testString),
			index:    30,
			expected: -1,
		},
		{
			name:     "middle",
			data:     []byte(testString),
			index:    7,
			expected: 7,
		},
		{
			name:     "end",
			data:     []byte(testString),
			index:    8,
			expected: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pos := ByteIndex(test.data, test.index)
			assert.Equal(t, test.expected, pos)
		})
	}
}

func TestByteIndexEnd(t *testing.T) {
	testString := "Hello, 世界"

	tests := []struct {
		name     string
		data     []byte
		index    int
		expected int
	}{
		{
			name:     "start",
			data:     []byte(testString),
			index:    0,
			expected: 1,
		},
		{
			name:     "second rune",
			data:     []byte(testString),
			index:    1,
			expected: 2,
		},
		{
			name:     "out of bounds",
			data:     []byte(testString),
			index:    30,
			expected: -1,
		},
		{
			name:     "middle",
			data:     []byte(testString),
			index:    7,
			expected: 10,
		},
		{
			name:     "end",
			data:     []byte(testString),
			index:    8,
			expected: 13,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pos := ByteIndexEnd(test.data, test.index)
			assert.Equal(t, test.expected, pos)
		})
	}
}

func TestRuneIndex(t *testing.T) {
	testString := "Hello, 世界"

	tests := []struct {
		name     string
		data     []byte
		index    int
		expected int
	}{
		{
			name:     "start",
			data:     []byte(testString),
			index:    0,
			expected: 0,
		},
		{
			name:     "second rune",
			data:     []byte(testString),
			index:    1,
			expected: 1,
		},
		{
			name:     "out of bounds",
			data:     []byte(testString),
			index:    30,
			expected: -1,
		},
		{
			name:     "middle",
			data:     []byte(testString),
			index:    7,
			expected: 7,
		},
		{
			name:     "end",
			data:     []byte(testString),
			index:    10,
			expected: 8,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pos := RuneIndex(test.data, test.index)
			assert.Equal(t, test.expected, pos)
		})
	}
}

func TestRuneLen(t *testing.T) {
	s := []struct {
		data     []byte
		index    int
		expected int
	}{
		{
			data:     []byte("Hello, 世界"),
			index:    1,
			expected: 1,
		},
		{
			data:     []byte("Hello, 世界"),
			index:    7,
			expected: 3,
		},
		{
			data:     []byte("Hello, 世界"),
			index:    8,
			expected: 3,
		},
	}

	for _, d := range s {
		actual := RuneLen(d.data, d.index)
		assert.Equal(t, d.expected, actual)
	}
}

func TestRune(t *testing.T) {
	s := []struct {
		data     []byte
		index    int
		expected rune
	}{
		{
			data:     []byte("Hello, 世界"),
			index:    7,
			expected: '世',
		},
	}

	for _, d := range s {
		actual := Rune(d.data, d.index)
		assert.Equal(t, d.expected, actual)
	}
}

func TestCutStart(t *testing.T) {
	s := []struct {
		data     []byte
		index    int
		expected []byte
	}{
		{
			data:     []byte("Hello, 世界"),
			index:    8,
			expected: []byte("界"),
		},
		{
			data:     []byte("Hello, 世界"),
			index:    9,
			expected: []byte(""),
		},
		{
			data:     []byte("Hello, 世界"),
			index:    0,
			expected: []byte("Hello, 世界"),
		},
	}

	for _, d := range s {
		actual := CutStart(d.data, d.index)
		assert.Equal(t, d.expected, actual)
	}
}

func TestCutEnd(t *testing.T) {
	s := []struct {
		data     []byte
		index    int
		expected []byte
	}{
		{
			data:     []byte("Hello, 世界"),
			index:    4,
			expected: []byte("Hell"),
		},
		{
			data:     []byte("Hello, 世界"),
			index:    9,
			expected: []byte("Hello, 世界"),
		},
		{
			data:     []byte("Hello, 世界"),
			index:    0,
			expected: []byte(""),
		},
	}

	for _, d := range s {
		actual := CutEnd(d.data, d.index)
		assert.Equal(t, d.expected, actual)
	}
}

func TestCutRange(t *testing.T) {
	s := []struct {
		data     []byte
		start    int
		end      int
		expected []byte
	}{
		{
			data:     []byte("Hello, 世界"),
			start:    1,
			end:      8,
			expected: []byte("ello, 世"),
		},
		{
			data:     []byte("Hello, 世界"),
			start:    9,
			end:      9,
			expected: []byte(""),
		},
		{
			data:     []byte("Hello, 世界"),
			start:    0,
			end:      0,
			expected: []byte(""),
		},
	}

	for _, d := range s {
		actual := CutRange(d.data, d.start, d.end)
		assert.Equal(t, d.expected, actual)
	}
}

func TestAppend(t *testing.T) {
	s := []struct {
		data     []byte
		data2    []byte
		expected []byte
	}{
		{
			data:     []byte("Hello, 世"),
			data2:    []byte("界"),
			expected: []byte("Hello, 世界"),
		},
	}

	for _, d := range s {
		actual := Append(d.data, d.data2...)
		assert.Equal(t, d.expected, actual)
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		index    int
		data2    []byte
		expected []byte
	}{
		{
			name:     "insert at the beginning",
			data:     []byte("Hllo, 世界"),
			index:    1,
			data2:    []byte("e"),
			expected: []byte("Hello, 世界"),
		},
		{
			name:     "insert in the middle",
			data:     []byte("Hello, 界"),
			index:    7,
			data2:    []byte("世"),
			expected: []byte("Hello, 世界"),
		},
		{
			name:     "insert in the end",
			data:     []byte("Hello, 世"),
			index:    8,
			data2:    []byte("界"),
			expected: []byte("Hello, 世界"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Insert(test.data, test.index, test.data2...)
			assert.Equal(t, string(test.expected), string(actual))
		})
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		from     int
		to       int
		data2    []byte
		expected []byte
	}{
		{
			name:     "replace at the beginning",
			data:     []byte("Hello, 世界"),
			from:     7,
			to:       7,
			data2:    []byte("界"),
			expected: []byte("Hello, 界界"),
		},
		{
			name:     "replace in the middle",
			data:     []byte("Hello, 世界"),
			from:     7,
			to:       7,
			data2:    []byte("界"),
			expected: []byte("Hello, 界界"),
		},
		{
			name:     "replace at the end",
			data:     []byte("Hello, 世界"),
			from:     8,
			to:       8,
			data2:    []byte("a"),
			expected: []byte("Hello, 世a"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Replace(test.data, test.from, test.to, test.data2...)
			assert.Equal(t, string(test.expected), string(actual))
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		start    int
		end      int
		expected []byte
	}{
		{
			name:     "delete at the beginning",
			data:     []byte("Hello, 世界"),
			start:    0,
			end:      0,
			expected: []byte("ello, 世界"),
		},
		{
			name:     "delete in the middle",
			data:     []byte("Hello, 世界"),
			start:    7,
			end:      7,
			expected: []byte("Hello, 界"),
		},
		{
			name:     "delete at the end",
			data:     []byte("Hello, 世界"),
			start:    8,
			end:      8,
			expected: []byte("Hello, 世"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := Delete(test.data, test.start, test.end)
			assert.Equal(t, string(test.expected), string(actual))
		})
	}
}
