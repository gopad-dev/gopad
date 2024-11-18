package file

import (
	"log"
	"slices"

	"github.com/charmbracelet/lipgloss/v2"

	"go.gopad.dev/gopad/gopad/editor/buffer"

	"go.gopad.dev/gopad/gopad/ls"
)

func (d *Document) SetDiagnostic(dType ls.DiagnosticType, version uint64, diagnostics []ls.Diagnostic) {
	// ignore outdated diagnostics
	if version < d.diagnosticVersions[dType] {
		log.Printf("skipping outdated diagnostics: %d < %d", version, d.diagnosticVersions[dType])
		return
	}

	// if we have a new version of diagnostics, update the version
	if version > d.diagnosticVersions[dType] {
		d.diagnosticVersions[dType] = version
	}

	// always clear diagnostics of this type
	d.ClearDiagnosticsByType(dType)

	// add new diagnostics
	d.Diagnostics = append(d.Diagnostics, diagnostics...)
}

func (d *Document) ClearDiagnosticsByType(dType ls.DiagnosticType) {
	d.Diagnostics = slices.DeleteFunc(d.Diagnostics, func(diag ls.Diagnostic) bool {
		return diag.Type == dType
	})
}

func (d *Document) DiagnosticsForLineCol(row int, col int) []ls.Diagnostic {
	p := buffer.Point{Row: row, Col: col}

	var diagnostics []ls.Diagnostic
	for _, diag := range d.Diagnostics {
		if diag.Range.Contains(p) {
			diagnostics = append(diagnostics, diag)
		}
	}
	return diagnostics
}

func (d *Document) HighestLineDiagnostic(row int) (ls.Diagnostic, int) {
	var (
		diagnostic ls.Diagnostic
		index      int
	)
	for i, diag := range d.Diagnostics {
		if diag.Range.ContainsRow(row) && (diagnostic.Severity == 0 || (diag.Severity < diagnostic.Severity || (diag.Severity <= diagnostic.Severity && diag.Priority > diagnostic.Priority))) {
			diagnostic = diag
			index = i
		}
	}
	return diagnostic, index
}

func (d *Document) HighestLineColDiagnostic(row int, col int) ls.Diagnostic {
	p := buffer.Point{Row: row, Col: col}

	var diagnostic ls.Diagnostic
	for _, diag := range d.Diagnostics {
		if diag.Range.Contains(p) && (diagnostic.Severity == 0 || (diag.Severity < diagnostic.Severity || (diag.Severity <= diagnostic.Severity && diag.Priority > diagnostic.Priority))) {
			diagnostic = diag
		}
	}

	return diagnostic
}

func (d *Document) HighestLineColDiagnosticStyle(style lipgloss.Style, row int, col int) lipgloss.Style {
	p := buffer.Point{Row: row, Col: col}

	var diagnostic ls.Diagnostic
	for _, diag := range d.Diagnostics {
		if diag.Range.Contains(p) && (diag.Severity > diagnostic.Severity || (diag.Severity >= diagnostic.Severity && diag.Priority > diagnostic.Priority)) {
			diagnostic = diag
		}
	}

	if diagnostic.Severity == 0 {
		return style
	}

	return diagnostic.Severity.CharStyle().Inherit(style)
}
