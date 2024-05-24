package lsp

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
)

func UpdateFileDiagnostics(name string, version int32, diagnostics []Diagnostic) tea.Cmd {
	return func() tea.Msg {
		return UpdateFileDiagnosticsMsg{
			Name:        name,
			Version:     version,
			Diagnostics: diagnostics,
		}
	}
}

type UpdateFileDiagnosticsMsg struct {
	Name        string
	Version     int32
	Diagnostics []Diagnostic
}

type DiagnosticSeverity int

const (
	DiagnosticSeverityNone        DiagnosticSeverity = 0
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

func (d DiagnosticSeverity) String() string {
	switch d {
	case DiagnosticSeverityNone:
		return "None"
	case DiagnosticSeverityError:
		return "Error"
	case DiagnosticSeverityWarning:
		return "Warning"
	case DiagnosticSeverityInformation:
		return "Information"
	case DiagnosticSeverityHint:
		return "Hint"
	}
	return "Unknown"
}

func (d DiagnosticSeverity) Icon() string {
	switch d {
	case DiagnosticSeverityError:
		return string(config.Theme.Icons.Error)
	case DiagnosticSeverityWarning:
		return string(config.Theme.Icons.Warning)
	case DiagnosticSeverityInformation:
		return string(config.Theme.Icons.Information)
	case DiagnosticSeverityHint:
		return string(config.Theme.Icons.Hint)
	}
	return " "
}

func (d DiagnosticSeverity) Style() lipgloss.Style {
	switch d {
	case DiagnosticSeverityError:
		return config.Theme.Editor.Diagnostics.ErrorStyle
	case DiagnosticSeverityWarning:
		return config.Theme.Editor.Diagnostics.WarningStyle
	case DiagnosticSeverityInformation:
		return config.Theme.Editor.Diagnostics.InformationStyle
	case DiagnosticSeverityHint:
		return config.Theme.Editor.Diagnostics.HintStyle
	}
	return lipgloss.NewStyle()
}

func (d DiagnosticSeverity) CharStyle() lipgloss.Style {
	switch d {
	case DiagnosticSeverityError:
		return config.Theme.Editor.Diagnostics.ErrorCharStyle
	case DiagnosticSeverityWarning:
		return config.Theme.Editor.Diagnostics.WarningCharStyle
	case DiagnosticSeverityInformation:
		return config.Theme.Editor.Diagnostics.InformationCharStyle
	case DiagnosticSeverityHint:
		return config.Theme.Editor.Diagnostics.HintCharStyle
	}
	return lipgloss.NewStyle()
}

type DiagnosticType int

const (
	DiagnosticTypeTreeSitter DiagnosticType = iota + 1
	DiagnosticTypeLanguageServer
)

func (d DiagnosticType) String() string {
	switch d {
	case DiagnosticTypeTreeSitter:
		return "TreeSitter"
	case DiagnosticTypeLanguageServer:
		return "LanguageServer"
	}
	return "Unknown"
}

type Diagnostic struct {
	Type            DiagnosticType
	Source          string
	Range           buffer.Range
	Severity        DiagnosticSeverity
	Code            string
	CodeDescription string
	Message         string
	Data            any
	Priority        int
}

func (d Diagnostic) ShortView() string {
	return d.Severity.Style().Render(fmt.Sprintf("%s %s", d.Severity.Icon(), strings.SplitN(d.Message, "\n", 2)[0]))
}

func (d Diagnostic) View(width int, height int) string {
	width = min(width, 80)
	height = max(height-2, 0)

	return config.Theme.Editor.Documentation.Style.
		Width(width).
		MaxHeight(height).
		Render(fmt.Sprintf("%s: %s %s\n\n%s\n%s", d.Type, d.Source, d.Severity.Style().Render(d.Severity.Icon()+" "), d.Message, d.Source))
}
