package file

import (
	"slices"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/ls"
)

func (f *File) SetDiagnostic(dType ls.DiagnosticType, version int32, diagnostics []ls.Diagnostic) {
	// ignore outdated diagnostics
	if version < f.diagnosticVersions[dType] {
		return
	}

	// if we have a new version of diagnostics, update the version
	if version > f.diagnosticVersions[dType] {
		f.diagnosticVersions[dType] = version
	}

	// always clear diagnostics of this type
	f.ClearDiagnosticsByType(dType)

	// add new diagnostics
	f.diagnostics = append(f.diagnostics, diagnostics...)
}

func (f *File) Diagnostics() []ls.Diagnostic {
	return f.diagnostics
}

func (f *File) ClearDiagnosticsByType(dType ls.DiagnosticType) {
	f.diagnostics = slices.DeleteFunc(f.diagnostics, func(diag ls.Diagnostic) bool {
		return diag.Type == dType
	})
}

func (f *File) DiagnosticsForLineCol(row int, col int) []ls.Diagnostic {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostics []ls.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) {
			diagnostics = append(diagnostics, diag)
		}
	}
	return diagnostics
}

func (f *File) HighestLineDiagnostic(row int) ls.Diagnostic {
	var diagnostic ls.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.ContainsRow(row) && (diagnostic.Severity == 0 || (diag.Severity < diagnostic.Severity || (diag.Severity <= diagnostic.Severity && diag.Priority > diagnostic.Priority))) {
			diagnostic = diag
		}
	}
	return diagnostic
}

func (f *File) HighestLineColDiagnostic(row int, col int) ls.Diagnostic {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostic ls.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) && (diagnostic.Severity == 0 || (diag.Severity < diagnostic.Severity || (diag.Severity <= diagnostic.Severity && diag.Priority > diagnostic.Priority))) {
			diagnostic = diag
		}
	}

	return diagnostic
}

func (f *File) HighestLineColDiagnosticStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostic ls.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) && (diag.Severity > diagnostic.Severity || (diag.Severity >= diagnostic.Severity && diag.Priority > diagnostic.Priority)) {
			diagnostic = diag
		}
	}

	if diagnostic.Severity == 0 {
		return style
	}

	return diagnostic.Severity.CharStyle().Inherit(style)
}

func (f *File) ShowCurrentDiagnostic() {
	f.showCurrentDiagnostic = true
}

func (f *File) HideCurrentDiagnostic() {
	f.showCurrentDiagnostic = false
}
