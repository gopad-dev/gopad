package buffer

import (
	"unicode/utf8"

	"go.gopad.dev/gopad/internal/xbytes"
)

var (
	lineEndingCRLF = []byte("\r\n")
	lineEndingLF   = []byte("\n")
)

type LineEnding int

const (
	LineEndingAuto LineEnding = iota
	LineEndingLF
	LineEndingCRLF
)

func (l LineEnding) String() string {
	switch l {
	case LineEndingAuto:
		return "Auto"
	case LineEndingLF:
		return "LF"
	case LineEndingCRLF:
		return "CRLF"
	}
	return "Unknown"
}

func (l LineEnding) Bytes() []byte {
	switch l {
	case LineEndingCRLF:
		return lineEndingCRLF
	default:
		return lineEndingLF
	}
}

func NewLine(data []byte) Line {
	return Line{
		data: data,
	}
}

func NewEmptyLine() Line {
	return Line{
		data: nil,
	}
}

type Line struct {
	data []byte
}

// Len returns the number of runes in the line (not bytes) without the line ending.
func (l Line) Len() int {
	return utf8.RuneCount(l.data)
}

// LenBytes returns the number of bytes in the line without the line ending.
func (l Line) LenBytes() int {
	return len(l.data)
}

func (l Line) Bytes() []byte {
	return l.data
}

func (l Line) String() string {
	return string(l.data)
}

func (l Line) Runes() []rune {
	return xbytes.Runes(l.data)
}

func (l Line) RunesRange(start int, end int) []rune {
	return xbytes.RunesRange(l.data, start, end)
}

func (l Line) Rune(index int) rune {
	return xbytes.Rune(l.data, index)
}

func (l Line) RuneBytes(index int) []byte {
	return []byte(string(xbytes.Rune(l.data, index)))
}

func (l Line) RunesBytesRange(start int, end int) []byte {
	return []byte(string(xbytes.RunesRange(l.data, start, end)))
}

func (l Line) RuneString(index int) string {
	return string(xbytes.Rune(l.data, index))
}

func (l Line) StringRange(start int, end int) string {
	return string(xbytes.CutRange(l.data, start, end))
}

func (l Line) RuneStrings() []string {
	runes := xbytes.Runes(l.data)
	strs := make([]string, len(runes))
	for i, r := range runes {
		strs[i] = string(r)
	}
	return strs
}

func (l Line) RuneIndex(index int) int {
	return xbytes.RuneIndex(l.data, index)
}

func (l Line) CutStart(index int) Line {
	l.data = xbytes.CutStart(l.data, index)
	return l
}

func (l Line) CutEnd(index int) Line {
	l.data = xbytes.CutEnd(l.data, index)
	return l
}

func (l Line) CutRange(start int, end int) Line {
	l.data = xbytes.CutRange(l.data, start, end)
	return l
}

func (l Line) Append(line Line) Line {
	l.data = xbytes.Append(l.data, line.data...)
	return l
}

func (l Line) AppendLines(lines ...Line) Line {
	for _, line := range lines {
		l.data = xbytes.Append(l.data, line.data...)
	}
	return l
}

func (l Line) Insert(index int, b ...byte) Line {
	l.data = xbytes.Insert(l.data, index, b...)
	return l
}

func (l Line) Replace(index int, b ...byte) Line {
	l.data = xbytes.Replace(l.data, index, b...)
	return l
}

func (l Line) ReplaceRange(start int, end int, b ...byte) Line {
	l.data = xbytes.ReplaceRange(l.data, start, end, b...)
	return l
}

func (l Line) Copy() Line {
	data := make([]byte, len(l.data))
	copy(data, l.data)
	return Line{
		data: data,
	}
}
