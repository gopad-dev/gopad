package editor

import (
	"go.gopad.dev/gopad/gopad/lsp"
)

func (f *File) SetInlayHint(hints []lsp.InlayHint) {
	f.inlayHints = hints
}

func (f *File) InlayHints() []lsp.InlayHint {
	return f.inlayHints
}

func (f *File) ClearInlayHints() {
	f.inlayHints = nil
}

func (f *File) InlayHintsForLineCol(row int, col int) []lsp.InlayHint {
	var hints []lsp.InlayHint
	for _, hint := range f.inlayHints {
		if hint.Position.Row == row && hint.Position.Col == col {
			hints = append(hints, hint)
		}
	}
	return hints
}

func (f *File) InlayHintsForLine(row int) []lsp.InlayHint {
	var hints []lsp.InlayHint
	for _, hint := range f.inlayHints {
		if hint.Position.Row == row {
			hints = append(hints, hint)
		}
	}
	return hints
}
