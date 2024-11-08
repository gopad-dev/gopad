package file

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
)

func (f *File) HighestMatchStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	var (
		currentStyle   *lipgloss.Style
		referenceStyle *lipgloss.Style
	)
	for _, match := range f.MatchesForLineCol(row, col) {
		if match.ReferenceType != "" {
			newStyle := getMatchingStyle(match.ReferenceType, f.Language.Name)
			if newStyle != nil {
				referenceStyle = newStyle
			}
			continue
		}

		newStyle := getMatchingStyle(match.Type, f.Language.Name)
		if newStyle != nil {
			currentStyle = newStyle
		}
	}

	if referenceStyle != nil {
		return style.Inherit(*referenceStyle)
	}

	if currentStyle != nil {
		return style.Inherit(*currentStyle)
	}

	return style
}

func getMatchingStyle(matchType string, name string) *lipgloss.Style {
	var currentStyle *lipgloss.Style

	for {
		codeStyle, ok := config.Theme.CodeStyles[fmt.Sprintf("%s.%s", matchType, name)]
		if ok {
			currentStyle = &codeStyle
			break
		}
		codeStyle, ok = config.Theme.CodeStyles[matchType]
		if ok {
			currentStyle = &codeStyle
			break
		}
		lastDot := strings.LastIndex(matchType, ".")
		if lastDot == -1 {
			break
		}
		matchType = matchType[:lastDot]
	}

	return currentStyle
}
