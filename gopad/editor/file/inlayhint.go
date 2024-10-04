package file

import (
	"log"

	"go.gopad.dev/gopad/gopad/ls"
)

func (f *File) SetInlayHint(version int32, hints []ls.InlayHint) {
	if version < f.inlayHintsVersion {
		log.Printf("skipping outdated inlay hints: %d < %d", version, f.inlayHintsVersion)
		return
	}
	if version > f.inlayHintsVersion {
		f.inlayHintsVersion = version
	}
	f.inlayHints = hints
}

func (f *File) InlayHints() []ls.InlayHint {
	return f.inlayHints
}

func (f *File) ClearInlayHints() {
	f.inlayHints = nil
}

func (f *File) InlayHintsForLineCol(row int, col int) []ls.InlayHint {
	var hints []ls.InlayHint
	for _, hint := range f.inlayHints {
		if hint.Position.Row == row && hint.Position.Col == col {
			hints = append(hints, hint)
		}
	}
	return hints
}

func (f *File) InlayHintsForLine(row int) []ls.InlayHint {
	var hints []ls.InlayHint
	for _, hint := range f.inlayHints {
		if hint.Position.Row == row {
			hints = append(hints, hint)
		}
	}
	return hints
}
