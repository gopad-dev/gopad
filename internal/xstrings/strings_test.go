package xstrings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCutNewlines(t *testing.T) {
	data := []struct {
		input    string
		expected string
	}{
		{"test\n", "test"},
		{"test\ntest", "test"},
		{"test\n\n", "test"},
	}

	for _, d := range data {
		actual := CutNewlines(d.input)
		assert.Equal(t, d.expected, actual)
	}
}
