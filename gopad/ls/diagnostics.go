package ls

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/config"
)

func UpdateFileDiagnostic(name string, dType DiagnosticType, version int32, diagnostics []Diagnostic) tea.Cmd {
	return func() tea.Msg {
		return UpdateFileDiagnosticMsg{
			Name:        name,
			Type:        dType,
			Version:     version,
			Diagnostics: diagnostics,
		}
	}
}

type UpdateFileDiagnosticMsg struct {
	Name        string
	Type        DiagnosticType
	Version     int32
	Diagnostics []Diagnostic
}

type DiagnosticSeverity int

const (
	DiagnosticSeverityNone    DiagnosticSeverity = 0
	DiagnosticSeverityError   DiagnosticSeverity = 1
	DiagnosticSeverityWarning DiagnosticSeverity = 2
	DiagnosticSeverityInfo    DiagnosticSeverity = 3
	DiagnosticSeverityHint    DiagnosticSeverity = 4
)

func (d DiagnosticSeverity) String() string {
	switch d {
	case DiagnosticSeverityNone:
		return "None"
	case DiagnosticSeverityError:
		return "Error"
	case DiagnosticSeverityWarning:
		return "Warning"
	case DiagnosticSeverityInfo:
		return "Info"
	case DiagnosticSeverityHint:
		return "Hint"
	}
	return "Unknown"
}

func (d DiagnosticSeverity) Icon() string {
	switch d {
	case DiagnosticSeverityError:
		return config.Theme.Icons.Error.Render()
	case DiagnosticSeverityWarning:
		return config.Theme.Icons.Warning.Render()
	case DiagnosticSeverityInfo:
		return config.Theme.Icons.Info.Render()
	case DiagnosticSeverityHint:
		return config.Theme.Icons.Hint.Render()
	}
	return " "
}

func (d DiagnosticSeverity) Style() lipgloss.Style {
	switch d {
	case DiagnosticSeverityError:
		return config.Theme.Diagnostic.ErrorStyle
	case DiagnosticSeverityWarning:
		return config.Theme.Diagnostic.WarningStyle
	case DiagnosticSeverityInfo:
		return config.Theme.Diagnostic.InfoStyle
	case DiagnosticSeverityHint:
		return config.Theme.Diagnostic.HintStyle
	}
	return lipgloss.NewStyle()
}

func (d DiagnosticSeverity) CharStyle() lipgloss.Style {
	switch d {
	case DiagnosticSeverityError:
		return config.Theme.Diagnostic.ErrorCharStyle
	case DiagnosticSeverityWarning:
		return config.Theme.Diagnostic.WarningCharStyle
	case DiagnosticSeverityInfo:
		return config.Theme.Diagnostic.InfoCharStyle
	case DiagnosticSeverityHint:
		return config.Theme.Diagnostic.HintCharStyle
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
		return "Grammar"
	case DiagnosticTypeLanguageServer:
		return "LanguageServer"
	}
	return "Unknown"
}

type Diagnostic struct {
	Type            DiagnosticType
	Name            string
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
	width = min(width, 60)
	height = max(height-2, 0)

	message := fmt.Sprintf("%s%s\n\n%s: %s", d.Severity.Style().Render(d.Severity.Icon()), config.Theme.UI.Documentation.Style.Inline(true).Render(" "+d.Message), d.Type, d.Name)
	if d.Source != "" {
		message += fmt.Sprintf(" - %s", d.Source)
	}
	if d.Code != "" {
		message += fmt.Sprintf(" - %s", d.Code)
	}
	if d.CodeDescription != "" {
		message += fmt.Sprintf(" (%s)", d.CodeDescription)
	}

	return config.Theme.UI.Documentation.Style.
		Width(width).
		MaxHeight(height).
		Render(message)
}
