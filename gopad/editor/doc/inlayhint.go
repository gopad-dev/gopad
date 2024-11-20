package doc

import (
	"log"

	"go.gopad.dev/gopad/gopad/ls"
)

func (d *Document) SetInlayHint(version uint64, hints []ls.InlayHint) {
	if version < d.inlayHintsVersion {
		log.Printf("skipping outdated inlay hints: %d < %d", version, d.inlayHintsVersion)
		return
	}
	if version > d.inlayHintsVersion {
		d.inlayHintsVersion = version
	}
	d.InlayHints = hints
}

func (d *Document) ClearInlayHints() {
	d.InlayHints = nil
}

func (d *Document) InlayHintsForLineCol(row int, col int) []ls.InlayHint {
	var hints []ls.InlayHint
	for _, hint := range d.InlayHints {
		if hint.Position.Row == row && hint.Position.Col == col {
			hints = append(hints, hint)
		}
	}
	return hints
}

func (d *Document) InlayHintsForLine(row int) []ls.InlayHint {
	var hints []ls.InlayHint
	for _, hint := range d.InlayHints {
		if hint.Position.Row == row {
			hints = append(hints, hint)
		}
	}
	return hints
}
