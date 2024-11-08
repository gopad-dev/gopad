package file

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
)

func getMatchingStyle(matchType string, languageName string) *lipgloss.Style {
	var currentStyle *lipgloss.Style

	for {
		codeStyle, ok := config.Theme.CodeStyles[fmt.Sprintf("%s.%s", matchType, languageName)]
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
