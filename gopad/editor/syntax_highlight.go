package editor

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
)

func (f *File) SetMatches(version int32, matches []Match) {
	if version < f.matchesVersion {
		return
	}
	if version > f.matchesVersion {
		f.matches = matches
		f.matchesVersion = version
		return
	}
	f.matches = append(f.matches, matches...)
}

func (f *File) Matches() []Match {
	return f.matches
}

func (f *File) MatchesForLineCol(row int, col int) []Match {
	pos := buffer.Position{Row: row, Col: col}

	var matches []Match
	for _, match := range f.matches {
		if match.Range.Contains(pos) {
			matches = append(matches, match)
		}
	}
	return matches
}

func (f *File) HighestMatchStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	for _, match := range f.MatchesForLineCol(row, col) {
		matchType := match.Type
		for {
			codeStyle, ok := config.Theme.Editor.CodeStyles[matchType]
			if ok {
				return style.Copy().Inherit(codeStyle)
			}
			lastDot := strings.LastIndex(matchType, ".")
			if lastDot == -1 {
				break
			}
			matchType = matchType[:lastDot]
		}
	}

	return style
}
