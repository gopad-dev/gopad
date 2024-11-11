package file

import (
	"log"

	"go.gopad.dev/gopad/gopad/ls"
)

func (f *Document) SetInlayHint(version uint64, hints []ls.InlayHint) {
	if version < f.inlayHintsVersion {
		log.Printf("skipping outdated inlay hints: %d < %d", version, f.inlayHintsVersion)
		return
	}
	if version > f.inlayHintsVersion {
		f.inlayHintsVersion = version
	}
	f.InlayHints = hints
}

func (f *Document) ClearInlayHints() {
	f.InlayHints = nil
}

func (f *Document) InlayHintsForLineCol(row int, col int) []ls.InlayHint {
	var hints []ls.InlayHint
	for _, hint := range f.InlayHints {
		if hint.Position.Row == row && hint.Position.Col == col {
			hints = append(hints, hint)
		}
	}
	return hints
}

func (f *Document) InlayHintsForLine(row int) []ls.InlayHint {
	var hints []ls.InlayHint
	for _, hint := range f.InlayHints {
		if hint.Position.Row == row {
			hints = append(hints, hint)
		}
	}
	return hints
}
