package editor

import (
	"slices"

	"github.com/charmbracelet/lipgloss"

	"go.gopad.dev/gopad/gopad/buffer"
	"go.gopad.dev/gopad/gopad/lsp"
)

func (f *File) SetDiagnostic(dType lsp.DiagnosticType, version int32, diagnostics []lsp.Diagnostic) {
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

func (f *File) Diagnostics() []lsp.Diagnostic {
	return f.diagnostics
}

func (f *File) ClearDiagnosticsByType(dType lsp.DiagnosticType) {
	f.diagnostics = slices.DeleteFunc(f.diagnostics, func(diag lsp.Diagnostic) bool {
		return diag.Type == dType
	})
}

func (f *File) DiagnosticsForLineCol(row int, col int) []lsp.Diagnostic {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostics []lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) {
			diagnostics = append(diagnostics, diag)
		}
	}
	return diagnostics
}

func (f *File) HighestLineDiagnostic(row int) lsp.Diagnostic {
	var diagnostic lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.ContainsRow(row) && (diagnostic.Severity == 0 || (diag.Severity < diagnostic.Severity || (diag.Severity <= diagnostic.Severity && diag.Priority > diagnostic.Priority))) {
			diagnostic = diag
		}
	}
	return diagnostic
}

func (f *File) HighestLineColDiagnostic(row int, col int) lsp.Diagnostic {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostic lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) && (diagnostic.Severity == 0 || (diag.Severity < diagnostic.Severity || (diag.Severity <= diagnostic.Severity && diag.Priority > diagnostic.Priority))) {
			diagnostic = diag
		}
	}

	return diagnostic
}

func (f *File) HighestLineColDiagnosticStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	pos := buffer.Position{Row: row, Col: col}

	var diagnostic lsp.Diagnostic
	for _, diag := range f.diagnostics {
		if diag.Range.Contains(pos) && (diag.Severity > diagnostic.Severity || (diag.Severity >= diagnostic.Severity && diag.Priority > diagnostic.Priority)) {
			diagnostic = diag
		}
	}

	if diagnostic.Severity == 0 {
		return style
	}

	return diagnostic.Severity.CharStyle().Copy().Inherit(style)
}

func (f *File) ShowCurrentDiagnostic() {
	f.showCurrentDiagnostic = true
}

func (f *File) HideCurrentDiagnostic() {
	f.showCurrentDiagnostic = false
}
