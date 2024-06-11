package editor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchGlobs(t *testing.T) {
	data := []struct {
		patterns []string
		path     string
		matches  bool
	}{
		{patterns: []string{"*.go"}, path: "main.go", matches: true},
		{patterns: []string{"*.go"}, path: "main.js", matches: false},
		{patterns: []string{"**/hypr/*.conf"}, path: "/home/user/.config/hypr/hyprland.conf", matches: true},
	}

	for _, d := range data {
		matches := matchGlobs(d.patterns, d.path)
		assert.Truef(t, matches == d.matches, "expected %s to match %v", d.path, d.patterns)
	}
}
