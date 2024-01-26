package xbytes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuneIndex(t *testing.T) {
	testString := "Hello, 世界"

	s := []struct {
		data        []byte
		pos         int
		expectedPos int
	}{
		{
			data:        []byte(testString),
			pos:         0,
			expectedPos: 0,
		},
		{
			data:        []byte(testString),
			pos:         1,
			expectedPos: 1,
		},
		{
			data:        []byte(testString),
			pos:         30,
			expectedPos: -1,
		},
		{
			data:        []byte(testString),
			pos:         0,
			expectedPos: 0,
		},
		{
			data:        []byte(testString),
			pos:         7,
			expectedPos: 7,
		},
		{
			data:        []byte(testString),
			pos:         8,
			expectedPos: 10,
		},
		{
			data:        []byte(testString),
			pos:         9,
			expectedPos: -1,
		},
	}

	for _, d := range s {
		pos := RuneIndex(d.data, d.pos)
		assert.Equal(t, d.expectedPos, pos)
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
	s := []struct {
		data     []byte
		index    int
		data2    []byte
		expected []byte
	}{
		{
			data:     []byte("Hllo, 世界"),
			index:    1,
			data2:    []byte("e"),
			expected: []byte("Hello, 世界"),
		},
		{
			data:     []byte("Hello, 界"),
			index:    7,
			data2:    []byte("世"),
			expected: []byte("Hello, 世界"),
		},
		{
			data:     []byte("Hello, 世"),
			index:    8,
			data2:    []byte("界"),
			expected: []byte("Hello, 世界"),
		},
	}

	for _, d := range s {
		actual := Insert(d.data, d.index, d.data2...)
		assert.Equal(t, d.expected, actual)
	}
}

func TestReplace(t *testing.T) {
	s := []struct {
		data     []byte
		i        int
		data2    []byte
		expected []byte
	}{
		{
			data:     []byte("Hello, 世界"),
			i:        7,
			data2:    []byte("界"),
			expected: []byte("Hello, 界界"),
		},
		{
			data:     []byte("Hello, 世界"),
			i:        8,
			data2:    []byte("a"),
			expected: []byte("Hello, 世a"),
		},
	}

	for _, d := range s {
		actual := Replace(d.data, d.i, d.data2...)
		assert.Equal(t, d.expected, actual)
	}
}
