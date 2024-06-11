package grammar

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/config"
)

var (
	msgStyle = lipgloss.NewStyle().MarginLeft(1).PaddingLeft(1).Border(lipgloss.Border{
		Left: ">",
	}, false, false, false, true)

	grammarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	infoStyle    = lipgloss.NewStyle().Faint(true)
)

func msgFromErr(err error) string {
	if err == nil {
		return ""
	}

	var msg string

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		msg = string(exitErr.Stderr)
	}

	if len(msg) == 0 {
		msg = err.Error()
	}

	return msgStyle.Render(msg) + "\n"
}

func filterLanguages(args []string, requireInstall bool) []config.LanguageConfig {
	var languages []config.LanguageConfig

	for name, language := range config.Languages.Languages {
		if len(args) > 0 && !slices.Contains(args, name) {
			continue
		}
		if language.Grammar == nil || (requireInstall && language.Grammar.Install == nil) {
			continue
		}

		languages = append(languages, language)
	}

	return languages
}

func ensureGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed: %w", err)
	}
	return nil
}

func ensureTreeSitterInstalled() error {
	_, err := exec.LookPath("tree-sitter")
	if err != nil {
		return fmt.Errorf("tree-sitter is not installed: %w", err)
	}
	return nil
}
