package buffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteLine_Append(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		data     string
		expected string
	}{
		{
			name:     "append to empty line",
			line:     "",
			data:     "hello",
			expected: "hello",
		},
		{
			name:     "append to non-empty line",
			line:     "world",
			data:     "hello",
			expected: "worldhello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).Append(NewLine([]byte(tt.data)))
			assert.Equal(t, tt.expected, line.String())
		})
	}
}

func TestByteLine_Prepend(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		data     string
		expected string
	}{
		{
			name:     "prepend to empty line",
			line:     "",
			data:     "hello",
			expected: "hello",
		},
		{
			name:     "prepend to non-empty line",
			line:     "world",
			data:     "hello",
			expected: "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).Prepend(NewLine([]byte(tt.data)))
			assert.Equal(t, tt.expected, line.String())
		})
	}
}

func TestByteLine_CutStart(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		index    int
		expected string
	}{
		{
			name:     "cut start",
			line:     "hello",
			index:    2,
			expected: "llo",
		},
		{
			name:     "cut start at end",
			line:     "hello",
			index:    5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).CutStart(tt.index)
			assert.Equal(t, tt.expected, line.String())
		})
	}
}

func TestByteLine_CutEnd(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		index    int
		expected string
	}{
		{
			name:     "cut end",
			line:     "hello",
			index:    2,
			expected: "he",
		},
		{
			name:     "cut end at start",
			line:     "hello",
			index:    0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).CutEnd(tt.index)
			assert.Equal(t, tt.expected, line.String())
		})
	}
}

func TestByteLine_CutRange(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		start    int
		end      int
		expected string
	}{
		{
			name:     "cut range",
			line:     "hello",
			start:    1,
			end:      3,
			expected: "el",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).CutRange(tt.start, tt.end)
			assert.Equal(t, tt.expected, line.String())
		})
	}
}

func TestByteLine_Insert(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		index    int
		data     string
		expected string
	}{
		{
			name:     "insert at start",
			line:     "world",
			index:    0,
			data:     "hello",
			expected: "helloworld",
		},
		{
			name:     "insert in middle",
			line:     "world",
			index:    2,
			data:     "hello",
			expected: "wohellorld",
		},
		{
			name:     "insert at end",
			line:     "world",
			index:    5,
			data:     "hello",
			expected: "worldhello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).Insert(tt.index, []byte(tt.data))
			assert.Equal(t, tt.expected, line.String())
		})
	}
}

func TestByteLine_Replace(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		start    int
		end      int
		data     string
		expected string
	}{
		{
			name:     "replace at start",
			line:     "world",
			start:    0,
			end:      1,
			data:     "hello",
			expected: "hellorld",
		},
		{
			name:     "replace in middle",
			line:     "world",
			start:    2,
			end:      3,
			data:     "hello",
			expected: "wohellod",
		},
		{
			name:     "replace at end",
			line:     "world",
			start:    3,
			end:      4,
			data:     "hello",
			expected: "worhello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).Replace(tt.start, tt.end, []byte(tt.data))
			assert.Equal(t, tt.expected, line.String())
		})
	}
}

func TestByteLine_Delete(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		start    int
		end      int
		expected string
	}{
		{
			name:     "delete at start",
			line:     "world",
			start:    0,
			end:      1,
			expected: "rld",
		},
		{
			name:     "delete in middle",
			line:     "world",
			start:    2,
			end:      4,
			expected: "wo",
		},
		{
			name:     "delete at end",
			line:     "world",
			start:    3,
			end:      4,
			expected: "wor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := NewLine([]byte(tt.line)).Delete(tt.start, tt.end)
			assert.Equal(t, tt.expected, line.String())
		})
	}
}
