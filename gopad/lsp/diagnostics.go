package lsp

import (
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
	default:
		return "Unknown"
	}
}

func (d DiagnosticSeverity) ShortString() string {
	return d.String()[:1]
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
	default:
		return lipgloss.NewStyle()
	}
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
	default:
		return lipgloss.NewStyle()
	}
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
	default:
		return "Unknown"
	}
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
