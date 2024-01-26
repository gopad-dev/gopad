package xrunes

import (
	"unicode"
	"unicode/utf8"
)

var replaceNewLine = []byte("\n")

func Sanitize(bytes []byte) []byte {
	dstBytes := make([]byte, 0, len(bytes))

	runes := []rune(string(bytes))
	for src := 0; src < len(runes); src++ {
		r := runes[src]
		switch {
		case r == utf8.RuneError:
			// skip

		case r == '\r' || r == '\n':
			if r == '\n' && src > 0 && runes[src-1] == '\r' {
				// Skip \n after \r.
				continue
			}
			dstBytes = append(dstBytes, replaceNewLine...)

		case r == '\t':
			// Keep tabs.
			dstBytes = append(dstBytes, []byte(string(runes[src]))...)

		case unicode.IsControl(r):
			// Other control characters: skip.

		default:
			// Keep the character.
			dstBytes = append(dstBytes, []byte(string(runes[src]))...)
		}
	}
	return dstBytes
}
